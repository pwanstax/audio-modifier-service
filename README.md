# Audio Modifier Backend

A user-friendly and powerful audio processing server with Voice Activity Detection (VAD) and Speech-to-Text (STT) features. Simply run the server using Docker and start processing your audio files in various formats, including .mp3, .wav, and .flac.

## Features

- Voice Activity Detection (VAD)
- Speech to Text (STT) using Google Cloud API
- Dockerized server for easy setup
- Supports .mp3, .wav, and .flac file formats

## Prerequisites

- Docker
- Google Cloud account with speech-to-text service activated

## Installation & Usage

1. Clone this repository.

2. Run the `buildnrun.sh` script to build and run the Docker container, The server will be available at http://localhost:8080.

3. Log in to your Google Cloud account by running: `gcloud init`
 
#### Note: Ensure that the speech-to-text service is activated for your Google Cloud account.

## API Endpoints

1. Get Token - GET
Endpoint: http://localhost:8080/getToken

No input required. Returns a generated token in the response body.

2. Upload Audio File - POST
Endpoint: http://localhost:8080/upload/audiofile

Requires a given token in the request header and an audio file in the request body. Returns information about available features.

3. Process File - GET
Endpoint: http://localhost:8080/upload/process

Requires a given token and user choice. Returns processed text.

## Supported File Types

- .mp3
- .wav
- .flac

## License

MIT
