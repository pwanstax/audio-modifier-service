FROM golang:1.16.5-buster

LABEL author="DEPRESSION_BACKEND"
LABEL git_commit=$GIT_COMMIT

WORKDIR /app

# install python and needed packages
RUN apt update
RUN apt -y upgrade

RUN apt -y install python3-pip

RUN pip3 install pandas
RUN pip3 install tqdm
RUN pip3 install webrtcvad

# install gcloud 
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg  add - && apt-get update -y && apt-get install google-cloud-sdk -y

# install ffmpeg
RUN apt install ffmpeg -y

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

ENV PORT 8080

ENV GIN_MODE release

RUN go build

RUN find . -name "*.go" -type f -delete

EXPOSE $PORT

CMD "./server"