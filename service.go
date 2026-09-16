package helloworld

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

type petpetService struct {
	client            *http.Client
	discordAPIBaseURL string
	petAPIBaseURL     string
}

func NewService(petApiBaseUrl string) *petpetService {
	return &petpetService{
		client:            &http.Client{Timeout: defaultPetpetHTTPTimeout, CheckRedirect: rejectRedirects},
		discordAPIBaseURL: "https://discord.com/api/v10",
		petAPIBaseURL:     petApiBaseUrl,
	}
}

func rejectRedirects(_ *http.Request, _ []*http.Request) error {
	return errors.New("redirects are not allowed")
}

func (service petpetService) deferInteraction(ctx context.Context, payload interaction, ephemeral bool) error {
	if payload.ID == "" || payload.Token == "" {
		return errors.New("Interaction is missing callback data")
	}
	data := map[string]any{}
	if ephemeral {
		data["flags"] = discordEphemeralFlag
	}
	body, err := json.Marshal(map[string]any{"type": responseDeferredChannelMessageWithSource, "data": data})
	if err != nil {
		return err
	}
	callbackURL := service.discordAPIBaseURL + "/interactions/" + url.PathEscape(payload.ID) + "/" + url.PathEscape(payload.Token) + "/callback"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, callbackURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := service.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Discord callback returned status %d", response.StatusCode)
	}
	return nil
}

func (service petpetService) petpetResponse(ctx context.Context, payload interaction, request petpetRequest) (responseMessage, []byte, string, string, error) {
	var (
		image       []byte
		filename    string
		contentType string
		content     string
		mentions    *allowedMentions
		err         error
	)

	switch request.subcommand {
	case "user":
		image, err = service.createUserPetpet(ctx, request.targetID, request.delay)
		filename, contentType = "petpet.webp", "image/webp"
		if request.mentions {
			authorID := interactionUserID(payload)
			if authorID == "" {
				return responseMessage{}, nil, "", "", errors.New("Interaction author is missing")
			}
			content = "<@" + authorID + "> has pet <@" + request.targetID + ">"
			mentions = &allowedMentions{Parse: []string{"users"}}
		}
	case "image":
		image, err = service.createImagePetpet(ctx, request.attachment, request.delay)
		filename, contentType = "petpet.gif", "image/gif"
	default:
		return responseMessage{}, nil, "", "", errors.New("Unknown petpet subcommand")
	}
	if err != nil {
		return responseMessage{}, nil, "", "", err
	}

	return responseMessage{Content: content, AllowedMentions: mentions}, image, filename, contentType, nil
}

func (service petpetService) createUserPetpet(ctx context.Context, userID string, delay int) ([]byte, error) {
	requestURL := service.petAPIBaseURL + "/ds/" + url.PathEscape(userID) + ".webp?delay=" + strconv.Itoa(delay)
	return service.fetchImage(ctx, http.MethodGet, requestURL, nil, "image/webp")
}

func (service petpetService) createImagePetpet(ctx context.Context, attachment discordAttachment, delay int) ([]byte, error) {
	image, err := service.fetchImage(ctx, http.MethodGet, attachment.URL, nil, "")
	if err != nil {
		return nil, err
	}
	contentType, ok := createAttachmentContentType(attachment.Filename)
	if !ok || !isExpectedImage(image, contentType) {
		return nil, errors.New("Image attachment content does not match its type")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, escapeFilename(attachment.Filename)))
	header.Set("Content-Type", attachment.ContentType)
	file, err := writer.CreatePart(header)
	if err != nil {
		return nil, err
	}
	if _, err := file.Write(image); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	requestURL := service.petAPIBaseURL + "/c?delay=" + strconv.Itoa(delay)
	return service.fetchImage(ctx, http.MethodPost, requestURL, &body, "image/gif", writer.FormDataContentType())
}

func isExpectedImage(image []byte, contentType string) bool {
	switch contentType {
	case "image/jpeg":
		return len(image) >= 3 && image[0] == 0xff && image[1] == 0xd8 && image[2] == 0xff
	case "image/png":
		return len(image) >= 8 && bytes.Equal(image[:8], []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n'})
	case "image/webp":
		return len(image) >= 12 && bytes.Equal(image[:4], []byte("RIFF")) && bytes.Equal(image[8:12], []byte("WEBP"))
	default:
		return false
	}
}

func (service petpetService) fetchImage(ctx context.Context, method, requestURL string, body io.Reader, expectedContentType string, contentType ...string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, err
	}
	if len(contentType) > 0 {
		request.Header.Set("Content-Type", contentType[0])
	}
	response, err := service.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Image service returned status %d", response.StatusCode)
	}
	if expectedContentType != "" {
		mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
		if err != nil || !strings.EqualFold(mediaType, expectedContentType) {
			return nil, errors.New("Image service returned an unexpected content type")
		}
	}
	return io.ReadAll(response.Body)
}

func (service petpetService) editError(ctx context.Context, payload interaction, content string) error {
	return service.editOriginal(ctx, payload, responseMessage{
		Content:         content,
		AllowedMentions: &allowedMentions{Parse: []string{}},
	}, "", "", nil)
}

func (service petpetService) editOriginal(ctx context.Context, payload interaction, message responseMessage, filename, contentType string, image []byte) error {
	if payload.ApplicationID == "" || payload.Token == "" {
		return errors.New("Interaction is missing webhook data")
	}
	payloadJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	payloadPart, err := writer.CreateFormField("payload_json")
	if err != nil {
		return err
	}
	if _, err := payloadPart.Write(payloadJSON); err != nil {
		return err
	}
	if len(image) > 0 {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="files[0]"; filename="%s"`, escapeFilename(filename)))
		header.Set("Content-Type", contentType)
		file, err := writer.CreatePart(header)
		if err != nil {
			return err
		}
		if _, err := file.Write(image); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}

	requestURL := service.discordAPIBaseURL + "/webhooks/" + url.PathEscape(payload.ApplicationID) + "/" + url.PathEscape(payload.Token) + "/messages/@original"
	request, err := http.NewRequestWithContext(ctx, http.MethodPatch, requestURL, &body)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := service.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Discord response update returned status %d", response.StatusCode)
	}
	return nil
}
