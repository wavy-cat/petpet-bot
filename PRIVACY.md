# Privacy Policy

*Revision from September 13, 2026.*

PetPet is a small, free service that turns a Discord avatar or an uploaded picture into a "petpet" GIF. It's maintained by one person, not a company, and it tries to collect as little as technically possible to keep the lights on.

## Definitions

- **Maintainer** — WavyCat, the individual developer who builds and operates PetPet. There is no company behind it.
- **Service** — the Website and the Bot, together.
- **Website** — the API available at `pet.wavycat.me`, including the Discord-avatar endpoint and the custom-upload endpoint. Its source code is public at [github.com/wavy-cat/petpet-go](https://github.com/wavy-cat/petpet-go) under the BSD-3-Clause license.
- **Bot** — the official PetPet Discord bot, which lets users request a petpet GIF directly from a Discord server by calling the Website's API on their behalf.
- **You** — anyone using the Website or the Bot.

## What the Service does with your data

### Discord-avatar requests

When you (or the Bot, on your behalf) request a GIF for a Discord user ID, the Website fetches that user's *public* avatar from Discord's own CDN and turns it into a GIF or WebP. No Discord login, token, or account access is required or requested — a user ID is all it takes, the same way it would be if you opened that avatar's URL directly.

### Custom uploads

If you upload your own image instead, it's read into memory, processed into a GIF/WebP, and handled the same way as any other generated result (see **Caching** below). Uploads are capped at 5&nbsp;MB and 1&nbsp;megapixel.

### The Bot

When you use a slash command, Discord's interaction system hands the Bot only what's needed to answer you: your Discord user ID, the target user's ID (if the command takes one), and the ID of the channel/server to reply in. The Bot does not read message content and does not request the privileged "message content" intent. It forwards the relevant IDs to the Website's API and posts the result back — it doesn't keep its own separate log of who ran what.

## Caching

To avoid hammering Discord's CDN and to speed up repeat requests, generated GIFs/WebPs (and, for custom uploads, the source image needed to produce them) are cached for **up to 3 days**, then automatically purged. Cache entries are keyed by request parameters (e.g. user ID, upload hash) — not by any account or identity.

## Logs

Cloudflare sits in front of the Website purely as a reverse proxy — there's no Workers code and no log export (Logpush/Instant Logs) configured, so the Maintainer doesn't receive or store any separate Cloudflare access logs. The only request logs that actually exist are the standard ones Google Cloud Run generates for the backend.

These Cloud Run logs contain routine technical metadata for every request: IP address (Cloudflare forwards the real visitor IP via the `CF-Connecting-IP` header, along with `CF-IPCountry` and a Cloudflare Ray ID for reference), User-Agent, request path including query parameters, timestamp, HTTP status code, response size, and processing time. They don't include uploaded images, generated GIFs, or other request bodies. Retention is set to **3 days**, after which logs are automatically deleted.

These logs aren't used for profiling or advertising, or shared with anyone beyond what's needed to run the underlying infrastructure (see **Third-party providers**).

## No analytics, no cookies, no ads

The Website and Bot don't run any analytics or tracking scripts, don't set cookies for tracking purposes, and don't show ads. There's nothing to sell your data to advertisers with, and we wouldn't anyway.

## Third-party providers

- **Cloudflare** — sits in front of the Website purely as a reverse proxy, providing DDoS/abuse protection. It's not used for logging, caching, or any application logic (no Workers). All traffic still passes through Cloudflare's network on its way to the backend, subject to [Cloudflare's Privacy Policy](https://www.cloudflare.com/privacypolicy/), but the Maintainer doesn't receive or store separate logs from Cloudflare's side.
- **Google Cloud Platform (Cloud Run)** — hosts and runs the Website's backend. See [Google Cloud's Privacy Notice](https://cloud.google.com/terms/cloud-privacy-notice).
- **Discord** — the source of the avatars fetched by the Discord-avatar endpoint, and the platform the Bot runs on. Discord's own infrastructure (for example, its link-preview crawler) may independently fetch URLs you share in Discord; that traffic is Discord's, not ours, and is covered by [Discord's Privacy Policy](https://discord.com/privacy), not this one.

Cloudflare, Google, and Discord are all U.S. companies, so data described above passes through U.S. infrastructure regardless of where you are.

## Data retention, at a glance

| What | Retention |
|---|---|
| Cached avatars/uploads & generated GIFs | 3 days |
| Cloud Run request logs | 3 days |

## Governing law and your rights

The Maintainer is an individual, not a registered company, and the Service isn't aimed at any particular country. Rather than claim a specific governing law, the Service is simply run on the principle of processing as little as possible, for as short a time as possible, and telling you exactly what that is.

If you believe the Service holds data about you that you'd like removed before it expires on its own (e.g., a cached custom upload), email the Maintainer and we'll act on reasonable requests. Depending on where you live, laws like the GDPR or CCPA may give you additional rights over that data; we'll do our best to honor them even though PetPet, by design, has very little to hand over.

## Children's privacy

The Service isn't directed at children under 13 (or the relevant minimum age in your jurisdiction) and doesn't knowingly collect data from them. There's no account system, so no one's identity is verified either way.

## Changes to this policy

This Policy may be updated from time to time as the Service changes. The current version is always the one published at this URL; if the document's history is tracked in a public repository, you can review past revisions there.

## Contact

Privacy questions, data-removal requests, or anything else: https://wavycat.me
