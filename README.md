# Discord Werewolf

A real-time game of werewolf on your Discord server! This is a work in progress clone of [Discord Werewolf Bot](https://github.com/CarterFiggins/werewolfBot).
Why a rewrite? The project has gotten a bit squirrelly and the lack of tests has made changes difficult. So we are 
rewriting in Go and using SQLite for maximum ease of deployment. 

This is currently in the alpha stage so there might be game disruption bugs. Currently only basic game functionality 
(wolves + villagers) is supported. When all current characters and game modes are fully supported this game will reach 
1.0 status. 

## Setup

### Discord Application

You will need to create a Discord application here: https://discord.com/developers/applications/

After creating a new application, go into its settings and:

* Get the application ID in "General Information" to use in the `CLIENT_ID` environment variable.
* In the "Bot" section:
    * Reset the token. Save the value for use in the `DISCORD_TOKEN` environment variable.
    * Give the bot all privileged gateway intents.
* In the "OAuth2" section:
    * Select the "bot" and "applications.commands" scopes under "OAuth2 URL Generator".
    * Select "Administrator" under "BOT PERMISSIONS".
    * Copy the generated URL below that and paste it into a new browser window.
    * Select the server you wish to invite the bot to.

### Running

You can either use the binaries provided in releases or the image on Docker Hub at `codebutter/discord-werewolf`. You
need to set `DISCORD_TOKEN` and `CLIENT_ID` environment variables with your Discord token and client ID, respectively.
The application takes the path to the database to use as the first argument. If the file doesn't exist it will be created. 
It is highly recommended to mount a volume at `/data` if you are using the container version.

Once the bot is connected to your guild, run the `/init` command to set up the game on your server.


## Further updates

As this is a work in progress documentation is rather sparse. Most of the base work for the game is completed and so we 
shouldn't see too much refactoring. Star this repo for updates!

 