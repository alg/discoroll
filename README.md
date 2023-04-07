# Discoroll

Acts like a tiny proxy converting Rollbar Webhook events into tidy Discord channel messages.

Currently supported:

- New item (new_item)
- 10^nth occurrence (exp_repeat_item)
- High velocity (item_velocity)
- Item reopened (reopened_item)
- Item resolved (resolved_item)

One installation can be reused by any number of projects and teams.

## Usage

- Build and deploy Discoroll. There's Makefile with `build` and `deploy` tasks. Deployment is configured for Fly.io.
- Register a webhook in your Discord Server configuration. You will get one in the form:
    
    https://discord.com/1093505700000000000/oW-p1l1VSUNR1JxpxGIZxMClsSrsiC4e7fZ-0-XXXXXXXXXXXXXXXXXXXXXXXXXXXXX

- Replace `discord.com` in the webhook URL with the host of where you deployed Discoroll. The rest of the URL remains the same.

## Contributions

File issues, discuss, suggest, you know the drill.

