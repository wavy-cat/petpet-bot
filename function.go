package helloworld

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("DiscordInteractions", discordInteractions)
}

func discordInteractions(w http.ResponseWriter, r *http.Request) {
	cfg := LoadConfig()

	w.Header().Set("Content-Type", "application/json")

	publicKeyHex := cfg.DiscordPublicKey
	if publicKeyHex == "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "DISCORD_PUBLIC_KEY is not configured"})
		return
	}

	publicKey, err := hex.DecodeString(publicKeyHex)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "DISCORD_PUBLIC_KEY is invalid"})
		return
	}

	signatureHex := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 || signatureHex == "" || timestamp == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Missing signature headers or request body"})
		return
	}
	if !validDiscordTimestamp(timestamp) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid request timestamp"})
		return
	}

	signature, err := hex.DecodeString(signatureHex)
	if err != nil || len(signature) != ed25519.SignatureSize || !ed25519.Verify(ed25519.PublicKey(publicKey), append([]byte(timestamp), body...), signature) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid request signature"})
		return
	}

	var payload interaction
	if err := json.Unmarshal(body, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		return
	}

	if payload.Type == interactionPing {
		writeJSON(w, http.StatusOK, map[string]int{"type": responsePong})
		return
	}

	petpetService := NewService(cfg.PetpetAPIBaseURL)

	if payload.Type == interactionApplicationCommand && payload.Data.Name == "petpet" {
		request, err := parsePetpetRequest(payload, cfg.MaxImageSize)
		if err != nil {
			writeJSON(w, http.StatusOK, interactionErrorResponse(err.Error(), false))
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), defaultPetpetHTTPTimeout)
		defer cancel()
		message, image, filename, contentType, err := petpetService.petpetResponse(ctx, payload, request)
		if err != nil {
			writeJSON(w, http.StatusOK, interactionErrorResponse("Unable to create the petpet image.", request.ephemeral))
			return
		}
		flags := 0
		if request.ephemeral {
			flags = discordEphemeralFlag
		}
		if err := writeMultipartInteractionResponse(w, responseMessage{
			Content:         message.Content,
			Flags:           flags,
			AllowedMentions: message.AllowedMentions,
			Attachments:     []responseAttachment{{ID: 0, Filename: filename}},
		}, filename, contentType, image); err != nil {
			return
		}
		return
	}

	writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown interaction type"})
}

func parsePetpetRequest(payload interaction, maxImageSize int64) (petpetRequest, error) {
	if len(payload.Data.Options) != 1 {
		return petpetRequest{}, errors.New("Invalid petpet subcommand")
	}

	subcommand := payload.Data.Options[0]
	if len(subcommand.Options) == 0 {
		return petpetRequest{}, errors.New("Invalid petpet subcommand")
	}
	request := petpetRequest{subcommand: subcommand.Name, delay: defaultPetpetDelay, mentions: true}
	seen := make(map[string]bool)

	for _, option := range subcommand.Options {
		if seen[option.Name] {
			return petpetRequest{}, errors.New("Duplicate petpet option")
		}
		seen[option.Name] = true
		if option.Value == nil {
			return petpetRequest{}, errors.New("Invalid petpet option")
		}

		switch subcommand.Name {
		case "user":
			switch option.Name {
			case "user":
				if json.Unmarshal(option.Value, &request.targetID) != nil || request.targetID == "" {
					return petpetRequest{}, errors.New("A user is required")
				}
			case "mentions":
				if json.Unmarshal(option.Value, &request.mentions) != nil {
					return petpetRequest{}, errors.New("Invalid mentions value")
				}
			case "speed":
				delay, err := parsePetpetDelay(option.Value)
				if err != nil {
					return petpetRequest{}, err
				}
				request.delay = delay
			case "ephemeral":
				if json.Unmarshal(option.Value, &request.ephemeral) != nil {
					return petpetRequest{}, errors.New("Invalid ephemeral value")
				}
			default:
				return petpetRequest{}, errors.New("Unknown petpet option")
			}
		case "image":
			if option.Name == "image" {
				var attachmentID string
				if json.Unmarshal(option.Value, &attachmentID) != nil || attachmentID == "" {
					return petpetRequest{}, errors.New("An image is required")
				}
				attachment, ok := payload.Data.Resolved.Attachments[attachmentID]
				if !ok {
					return petpetRequest{}, errors.New("Image attachment was not resolved")
				}
				if err := validateAttachment(attachment, maxImageSize); err != nil {
					return petpetRequest{}, err
				}
				request.attachment = attachment
				continue
			}
			switch option.Name {
			case "speed":
				delay, err := parsePetpetDelay(option.Value)
				if err != nil {
					return petpetRequest{}, err
				}
				request.delay = delay
			case "ephemeral":
				if json.Unmarshal(option.Value, &request.ephemeral) != nil {
					return petpetRequest{}, errors.New("Invalid ephemeral value")
				}
			default:
				return petpetRequest{}, errors.New("Unknown petpet option")
			}
		default:
			return petpetRequest{}, errors.New("Unknown petpet subcommand")
		}
	}

	if subcommand.Name == "user" && request.targetID == "" {
		return petpetRequest{}, errors.New("A user is required")
	}
	if subcommand.Name == "image" && request.attachment.URL == "" {
		return petpetRequest{}, errors.New("An image is required")
	}
	return request, nil
}

func parsePetpetDelay(value json.RawMessage) (int, error) {
	var delay int
	if json.Unmarshal(value, &delay) != nil {
		return 0, errors.New("Invalid speed value")
	}
	switch delay {
	case 2, 3, 5, 8:
		return delay, nil
	default:
		return 0, errors.New("Invalid speed value")
	}
}

func createAttachmentContentType(filename string) (string, bool) {
	contentType, ok := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
	}[strings.ToLower(filepath.Ext(filename))]
	return contentType, ok
}

func validateAttachment(attachment discordAttachment, maxImageSize int64) error {
	expectedContentType, ok := createAttachmentContentType(attachment.Filename)
	if !ok {
		return errors.New("Image must be a PNG, JPEG, or WebP file")
	}
	if attachment.ContentType != "" {
		mediaType, _, err := mime.ParseMediaType(attachment.ContentType)
		if err != nil || !strings.EqualFold(mediaType, expectedContentType) {
			return errors.New("Image must be a PNG, JPEG, or WebP file")
		}
	}
	if attachment.Size <= 0 || attachment.Size > maxImageSize {
		return errors.New("Image too big")
	}

	parsedURL, err := url.Parse(attachment.URL)
	if err != nil || parsedURL.Scheme != "https" || !isDiscordAttachmentHost(parsedURL.Hostname()) {
		return errors.New("Invalid image attachment URL")
	}
	return nil
}

func isDiscordAttachmentHost(host string) bool {
	host = strings.ToLower(host)
	return host == "cdn.discordapp.com" || host == "media.discordapp.net"
}

func validDiscordTimestamp(value string) bool {
	timestamp, err := strconv.ParseInt(value, 10, 64)
	if err != nil || timestamp <= 0 {
		return false
	}
	now := time.Now()
	requestTime := time.Unix(timestamp, 0)
	return now.Sub(requestTime) <= maxDiscordTimestampSkew && requestTime.Sub(now) <= maxDiscordTimestampSkew
}

func writeMultipartInteractionResponse(w http.ResponseWriter, message responseMessage, filename, contentType string, image []byte) error {
	message.Attachments = []responseAttachment{{ID: 0, Filename: filename}}
	payloadJSON, err := json.Marshal(map[string]any{
		"type": responseChannelMessageWithSource,
		"data": message,
	})
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
	fileHeader := make(textproto.MIMEHeader)
	fileHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="files[0]"; filename="%s"`, escapeFilename(filename)))
	fileHeader.Set("Content-Type", contentType)
	file, err := writer.CreatePart(fileHeader)
	if err != nil {
		return err
	}
	if _, err := file.Write(image); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	w.Header().Set("Content-Type", writer.FormDataContentType())
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body.Bytes())
	return err
}

func escapeFilename(filename string) string {
	filename = filepath.Base(filename)
	if filename == "." || filename == "\\" || filename == "" {
		return "image"
	}
	return strings.NewReplacer(`\`, "_", `"`, "_", "\r", "_", "\n", "_").Replace(filename)
}

func interactionUserID(payload interaction) string {
	if payload.Member != nil && payload.Member.User != nil {
		return payload.Member.User.ID
	}
	if payload.User != nil {
		return payload.User.ID
	}
	return ""
}

func interactionUserName(payload interaction) string {
	if payload.Member != nil && payload.Member.User != nil {
		return payload.Member.User.Username
	}
	if payload.User != nil {
		return payload.User.Username
	}
	return ""
}

func interactionErrorResponse(content string, ephemeral bool) map[string]any {
	flags := 0
	if ephemeral {
		flags = discordEphemeralFlag
	}
	return map[string]any{
		"type": responseChannelMessageWithSource,
		"data": map[string]any{
			"content": content,
			"flags":   flags,
			"allowed_mentions": map[string][]string{
				"parse": {},
			},
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
