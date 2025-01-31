# Shopware Copilot Extension

This repository contains a Copilot extension for Shopware. It allows you to ask questions about the Shopware codebase and developer docs.

To use it install the app first: https://github.com/apps/shopware and then it is available in the Copilot at https://github.com/copilot

## Development



Clone the Shopware Docs into the `data` directory:

and create a `.env` file with the following variables:

```
CLIENT_ID=
CLIENT_SECRET=
FQDN=<where-the-app-runs>
```

For the client id and client secret you need to create an app in your github account like:

1. In the `Copilot` tab of your Application settings (`https://github.com/settings/apps/<app_name>/agent`)
- Set the URL that was set for your FQDN above with the endpoint `/agent` (e.g. `https://6de513480979.ngrok.app/agent`)
- Set the Pre-Authorization URL with the endpoint `/auth/authorization` (e.g. `https://6de513480979.ngrok.app/auth/authorization`)
2. In the `General` tab of your application settings (`https://github.com/settings/apps/<app_name>`)
- Set the `Callback URL` with the `/auth/callback` endpoint (e.g. `https://6de513480979.ngrok.app/auth/callback`)
- Set the `Homepage URL` with the base ngrok endpoint (e.g. `https://6de513480979.ngrok.app/auth/callback`)
3. Ensure your permissions are enabled in `Permissions & events` > 
- `Account Permissions` > `Copilot Chat` > `Access: Read Only`
4. Ensure you install your application at (`https://github.com/apps/<app_name>`)


The local GITHUB_TOKEN must be right now dumped, by adding a log into the `service.go` file and using once the agent.

After that you can run the `index` command to embed all files in the `data` directory to the vector database.

Alternatively, you can download the `db.zip` from the release and unzip it into the root directory.

