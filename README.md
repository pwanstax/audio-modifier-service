# Depression App Backend

## Required
  - Only Docker

## Usage
``` sh
./buildnrun.sh
```
Server will run at http://localhost:8080

### to use speech to text service
### you have to login to your gcloud by running
``` sh
gcloud init
```
*** note that you speech-to-text service has to be activated ***

## Endpoint
### Upload Audio File - POST
http://localhost:8080/upload/audiofile

  Require given token in request header and audio file in body.
  Return information about feature (which feature is available).

### Process File - GET 
http://localhost:8080/upload/process
  
  Require given token and user choice.
  Return processed text.

### Get Token - GET
http://localhost:8080/getToken
  
  Require nothing. Return generated token in response body.

## Associated File Types
  - .mp3
  - .wav
  - .flac

## License
[MIT](https://choosealicense.com/licenses/mit/)
