package main

type StoredObjectRecord struct {
	S3 struct {
		Bucket struct {
			Name string
		}
		Object struct {
			Key string
		}
	}
}

type WebhookPayload struct {
	EventType string
	Records   []StoredObjectRecord
}

type Object struct {
	Bucket string
	Key    string
}

func (p *WebhookPayload) Object() Object {
	s3 := p.Records[0].S3

	return Object{
		Bucket: s3.Bucket.Name,
		Key:    s3.Object.Key,
	}
}

type PublisherConfig struct {
	CvmfsRepo        string
	MinioStoragePath string
}

type StatusRequest struct {
	Key    string
	Status string
}
