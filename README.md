# Trile

TRanslate fILE. Telegram bot written in go. Accepts messages with any
file types in them and converts them with LibreOffice API.

## Prerun

Create `data` dir in project root (along `.gitignore` file). It is as
temporary file storage.
```sh
cd trile
mkdir data
```

Create `.env` with `BOT_API_KEY` var - you TG bot api key
After creation you shoud see something like this
```sh
cat .env
# BOT_API_KEY=************
```

## Run

Dont forget to follow prerun steps first!

Run `main.go` file
```sh
go run main.go
# 2001/01/01 23:59:59 Starting LibreOffice background instance...
# 2001/01/01 23:59:59 LO started successfully
# 2001/01/01 23:59:59 Authorized on account @your_awesome_bot "https://t.me/your_awesome_bot"
```

Follow your bot url and try sending some `.pptx`, `.docx`, `.xlsx` etc.

Enjoy!

## License

Trile is granted under MIT license
