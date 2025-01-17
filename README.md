# Trile

TRanslate fILE. Telegram bot written in go. Accepts messages with any
file types in them and converts them with LibreOffice API.

## Prerun

Create `.env` with `BOT_API_KEY` - your TG bot api key
After creation you shoud see something like this
```sh
cat .env
# BOT_API_KEY=************
```

## Run

Dont forget to follow prerun steps first!

Build and run project
```sh
go build
./trile
# Logs will be stored in "/path/to/logs/trile.log"
# Authorized on account @your_awesome_bot "https://t.me/your_awesome_bot"
```

Follow your bot url and try sending some `.pptx`, `.docx`, `.xlsx` etc.

Enjoy!

## License

Trile is granted under MIT license
