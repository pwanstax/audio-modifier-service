package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"server/entity"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type ProcessService interface {
	ClearFile(format *entity.SttProcessFormat) error
	SetupClient(format *entity.SttProcessFormat) error
	SetupBucket(format *entity.SttProcessFormat, bucketName string)
	AsynchronousSTT(format *entity.SttProcessFormat, fileID string) (interface{}, error)
	VAD(token string, fileID string) error
}

type processService struct {
}

func NewProcessService() ProcessService {
	return &processService{}
}

func (service *processService) SetupClient(format *entity.SttProcessFormat) error {
	var err error
	format.Client, err = storage.NewClient(context.Background())
	if err != nil {
		return err

	}
	return nil

}

func (service *processService) openBucket(format *entity.SttProcessFormat) {
	format.Bucket = format.Client.Bucket(format.BucketName)
}

func (service *processService) SetupBucket(format *entity.SttProcessFormat, bucketName string) {
	format.BucketName = bucketName
	service.openBucket(format)
}

func (service *processService) resultRequest(processName, token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", "https://speech.googleapis.com/v1/operations/"+processName, nil)
	if err != nil {
		return nil, err

	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return req, nil

}

func (service *processService) resultMap(rawResponse *http.Response) (map[string]interface{}, error) {
	respondByte, _ := ioutil.ReadAll(rawResponse.Body)
	defer rawResponse.Body.Close()
	var resultResponse map[string]interface{}
	err := json.Unmarshal(respondByte, &resultResponse)
	if err != nil {
		return nil, err

	}
	return resultResponse, nil

}

func (service *processService) getResult(processName, token string) (map[string]interface{}, error) {
	resultRequest, err := service.resultRequest(processName, token)
	if err != nil {
		return nil, err

	}
	responseRaw, err := http.DefaultClient.Do(resultRequest)
	if err != nil {
		return nil, err

	}
	resultResponse, err := service.resultMap(responseRaw)
	if err != nil {
		return nil, err

	}
	return resultResponse, nil

}

func (service *processService) credentialsToken() string {
	fullPath, _ := os.Getwd()
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", fullPath+"/admin.json")
	cmd := exec.Command("gcloud", "auth", "application-default", "print-access-token")
	stdout, _ := cmd.CombinedOutput()
	token := "Bearer " + string(stdout)
	token = strings.Replace(token, "\r", "", 10)
	token = strings.Replace(token, "\n", "", 10)
	return token

}

func (service *processService) asynchronousSTTProcessBody(gcsURI string) *strings.Reader {
	body := strings.NewReader(`{
  		'config': {
    		'language_code': 'th-TH',
			'enable_word_time_offsets': 'true',
			'enable_automatic_punctuation': 'true',
			'audio_channel_count': 2,
			'enable_separate_recognition_per_channel': 'true',
			'diarization_config': {
				'enableSpeakerDiarization': 'true',
				'minSpeakerCount': 1,
				'maxSpeakerCount': 6
				}
  			},
  		'audio':{
    		'uri': '` + gcsURI + `'
  		}
	}`)
	return body

}

func (service *processService) sttProcessRequest(gcsURI string) (*http.Request, error) {
	token := service.credentialsToken()
	body := service.asynchronousSTTProcessBody(gcsURI)
	method := "POST"
	sttURL := "https://speech.googleapis.com/v1/speech:longrunningrecognize"
	req, err := http.NewRequest(method, sttURL, body)
	if err != nil {
		return nil, err

	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return req, nil

}

func (service *processService) curlSttProcessRequest(gcsURI string) (*http.Response, error) {
	processRequest, err := service.sttProcessRequest(gcsURI)
	if err != nil {
		return nil, err

	}
	httpRespond, err := http.DefaultClient.Do(processRequest)
	if err != nil {
		return nil, err

	}
	return httpRespond, nil

}

func (service *processService) operationName(rawResponse *http.Response) (string, error) {
	var response map[string]interface{}
	bodyBytes, _ := ioutil.ReadAll(rawResponse.Body)
	defer rawResponse.Body.Close()
	err := json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return "", err

	}
	name := fmt.Sprintf("%v", response["name"])
	return name, nil

}

func (service *processService) AsynchronousSTTFile(gcsURI string) (interface{}, error) {
	httpResponse, err := service.curlSttProcessRequest(gcsURI)
	if err != nil {
		return nil, err

	}
	name, err := service.operationName(httpResponse)
	if err != nil {
		return nil, err

	}
	var resultResponse map[string]interface{}
	for {
		credentialToken := service.credentialsToken()
		resultResponse, err = service.getResult(name, credentialToken)
		if err != nil {
			return nil, err

		}
		if fmt.Sprintf("%v", resultResponse["done"]) == "true" {
			break

		}
		time.Sleep(5 * time.Second)
	}
	result := resultResponse["response"]
	return result, nil

}

func (service *processService) fileName(cloudPath, secret string) string {
	return strings.Replace(cloudPath, secret+"/", "", 1)

}

func (service *processService) deleteAccomplishedFile(bucket *storage.BucketHandle, cloudPath string) error {
	object := bucket.Object(cloudPath)
	err := object.Delete(context.Background())
	if err != nil {
		return err

	}
	return nil

}

func (service *processService) AsynchronousSTT(format *entity.SttProcessFormat, fileID string) (interface{}, error) {
	var results map[string]interface{} = make(map[string]interface{})
	itr := format.Bucket.Objects(context.Background(), &storage.Query{Prefix: format.JwtSecret})
	for {
		fileAttrs, err := itr.Next()
		if err == iterator.Done {
			break

		}
		if err != nil {
			return nil, err

		}
		thisName := strings.Join(strings.Split(fileAttrs.Name, ".")[:len(strings.Split(fileAttrs.Name, "."))-1], ".")

		if thisName[len(thisName)-9:] == fileID {
			fileName := service.fileName(fileAttrs.Name, format.JwtSecret)
			fmt.Println(fileAttrs.Name)
			gcsURI := fmt.Sprintf("gs://%s/%s", format.BucketName, fileAttrs.Name)
			result, err := service.AsynchronousSTTFile(gcsURI)
			if err != nil {
				return nil, err

			}
			results[fileName] = result
			err = service.deleteAccomplishedFile(format.Bucket, fileAttrs.Name)
			if err != nil {
				return nil, err

			}
			return result, nil
		}
	}
	// if len(results) != 0 {
	// 	return results, nil
	// }
	return nil, nil

}

func (service *processService) ClearFile(format *entity.SttProcessFormat) error {
	itr := format.Bucket.Objects(context.Background(), &storage.Query{Prefix: format.JwtSecret})
	for {
		fileAttrs, err := itr.Next()
		if err == iterator.Done {
			break

		}
		if err != nil {
			return err

		}
		err = service.deleteAccomplishedFile(format.Bucket, fileAttrs.Name)
		if err != nil {
			return err

		}
	}

	err := service.deleteAccomplishedFile(format.Bucket, format.JwtSecret)
	if err != nil {
		return err

	}
	return nil
}

func (service *processService) VAD(secret string, fileID string) error {
	// os.Setenv("PATH", "/Library/Frameworks/Python.framework/Versions/3.9/bin:$PATH") //path to python3
	// cmd := exec.Command("python3.9", "./automate_vad_webrtcvad_users.py", "-s", secret)
	cmd := exec.Command("python3", "./automate_vad_webrtcvad_users.py", "--token", secret, "--fileid", fileID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
