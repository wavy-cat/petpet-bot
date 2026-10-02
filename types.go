package main

import (
	"encoding/json"
	"time"
)

const (
	interactionPing = iota + 1
	interactionApplicationCommand
	interactionMessageComponent
	interactionApplicationCommandAutocomplete
	interactionModalSubmit
)

const (
	responsePong                             = 1
	responseChannelMessageWithSource         = 4
	responseDeferredChannelMessageWithSource = 5
)

const (
	discordEphemeralFlag     = 1 << 6
	defaultPetpetDelay       = 3
	maxDiscordTimestampSkew  = 1 * time.Minute
	defaultPetpetHTTPTimeout = 30 * time.Second
)

type interaction struct {
	ID            string      `json:"id"`
	ApplicationID string      `json:"application_id"`
	Token         string      `json:"token"`
	Type          int         `json:"type"`
	Data          commandData `json:"data"`
	Member        *struct {
		User *discordUser `json:"user"`
	} `json:"member"`
	User *discordUser `json:"user"`
}

type commandData struct {
	Name     string          `json:"name"`
	Options  []commandOption `json:"options"`
	Resolved resolvedData    `json:"resolved"`
}

type commandOption struct {
	Name    string          `json:"name"`
	Value   json.RawMessage `json:"value"`
	Options []commandOption `json:"options"`
}

type resolvedData struct {
	Attachments map[string]discordAttachment `json:"attachments"`
}

type discordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type discordAttachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	URL         string `json:"url"`
}

type petpetRequest struct {
	subcommand string
	targetID   string
	attachment discordAttachment
	delay      int
	mentions   bool
	ephemeral  bool
}

type allowedMentions struct {
	Parse []string `json:"parse"`
}

type responseAttachment struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
}

type responseMessage struct {
	Content         string               `json:"content,omitempty"`
	Flags           int                  `json:"flags,omitempty"`
	AllowedMentions *allowedMentions     `json:"allowed_mentions,omitempty"`
	Attachments     []responseAttachment `json:"attachments,omitempty"`
}
