# RenameBot

A dead-simple bot to rename Telegram document files. It takes the document and the new filename, downloads the file, renames it, and uploads it back to you!

## Deployment:

1. Have a local-hosted [telegram-bot-api](https://github.com/tdlib/telegram-bot-api) instance available.

2. Clone this repo.

```
git clone https://github.com/haashemi/ReNamaeBot.git
```

3. Rename `config.example.yaml` to `config.yaml` and fill in the fields.

4. Build and run the bot!

```
go build .
./ReNamaeBot
```

## Usage

Just send your document to the bot and follow its instructions.
