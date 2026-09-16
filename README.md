# petpet-bot

Discord serverless bot for generating petpet images. Working on Cloud Run (Google Cloud).

You can install this bot to your account by clicking on the link: https://discord.com/oauth2/authorize?client_id=1470980091214434546

**This project was written almost entirely using AI, and it actually works.!**

## Register `/petpet`

Set the variables, then register the command globally:

```sh
export DISCORD_APPLICATION_ID="your-application-id"
export DISCORD_BOT_TOKEN="your-bot-token"

curl --fail-with-body --request POST \
  --url "https://discord.com/api/v10/applications/${DISCORD_APPLICATION_ID}/commands" \
  --header "Authorization: Bot ${DISCORD_BOT_TOKEN}" \
  --header "Content-Type: application/json" \
  --data '{
    "name": "petpet",
    "description": "Create a petpet animation",
    "type": 1,
    "integration_types": [0, 1],
    "contexts": [0, 1, 2],
    "options": [
      {
        "type": 1,
        "name": "user",
        "description": "Pet a Discord user",
        "options": [
          {
            "type": 6,
            "name": "user",
            "description": "User to pet",
            "required": true
          },
          {
            "type": 5,
            "name": "mentions",
            "description": "Mention the users in the response"
          },
          {
            "type": 4,
            "name": "speed",
            "description": "Animation speed",
            "choices": [
              {"name": "Fast", "value": 2},
              {"name": "Default", "value": 3},
              {"name": "Slow", "value": 5},
              {"name": "Slower", "value": 8}
            ]
          },
          {
            "type": 5,
            "name": "ephemeral",
            "description": "Only show the response to you"
          }
        ]
      },
      {
        "type": 1,
        "name": "image",
        "description": "Pet an uploaded image",
        "options": [
          {
            "type": 11,
            "name": "image",
            "description": "PNG, JPEG, or WebP image to pet",
            "required": true,
            "file_types": [".jpg", ".jpeg", ".png", ".webp"]
          },
          {
            "type": 4,
            "name": "speed",
            "description": "Animation speed",
            "choices": [
              {"name": "Fast", "value": 2},
              {"name": "Default", "value": 3},
              {"name": "Slow", "value": 5},
              {"name": "Slower", "value": 8}
            ]
          },
          {
            "type": 5,
            "name": "ephemeral",
            "description": "Only show the response to you"
          }
        ]
      }
    ]
  }'
```

For a guild-only test command, add `/guilds/${DISCORD_GUILD_ID}` between `applications/${DISCORD_APPLICATION_ID}` and `/commands` in the URL above.

`integration_types` controls where the app is installed: `0` is Guild Install and `1` is User Install. `contexts` controls where the command can be used: `0` is guild channels, `1` is a DM with the bot, and `2` is a DM or group DM with other users.

In the Discord Developer Portal, open **Installation**:

1. Under **Installation Contexts**, enable **User Install** and **Guild Install**.
2. Under **Default Install Settings → User Install**, add the `applications.commands` scope.
3. Under **Default Install Settings → Guild Install**, keep `applications.commands` and add `bot` if the app is also used in servers.
4. Copy the **Install Link**, open it while logged in, choose **Add to my apps**, and authorize the app. Existing installations do not automatically gain the new User Install context.

The global command registration does not install the application into a user's account; it only creates the command. If `petpet` still does not appear in a DM, reinstall the app using the updated install link and verify that the command has `integration_types: [0, 1]` and `contexts: [0, 1, 2]`.

Set `DISCORD_PUBLIC_KEY` in the deployed function environment to the Discord application public key. Never put a real bot token in this file; revoke the token shown in shell history or source control if it was exposed.
