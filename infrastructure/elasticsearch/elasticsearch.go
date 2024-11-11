package elasticsearch

import (
	"context"

	"github.com/olivere/elastic/v7"
	"go-quickstart/infrastructure/log"
)

var ctx = context.Background()

const (
	// mapping init es map struct
	mapping = `
{
   "settings":{
      "number_of_shards": 1,
      "number_of_replicas": 0 
   },
}`
	// index init es search-index
	index = "index"
)

type Client interface {
	Ping(esUrl string) (int, error)
	AddDoc(index string, data any) (bool, error)
}

func NewElasticSearchClient(esUrl string) (Client, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(esUrl),
		elastic.SetSniff(false),
	)
	if err != nil {
		log.GetTextLogger().Fatal(err.Error())
	}
	info, code, err := client.Ping(esUrl).Do(context.Background())
	if err != nil {
		log.GetTextLogger().Fatal("error is :%v", err)
	}
	log.GetTextLogger().Info("Elasticsearch call code:", code, " version:", info.Version.Number)

	initEsIndex(client)

	return &esClient{es: client}, nil
}

func initEsIndex(client *elastic.Client) {
	addIndex(client, index, mapping)
}

type esClient struct {
	es *elastic.Client
}

func addIndex(es *elastic.Client, index string, mapping string) bool {
	exists, _ := checkIndex(es, index)
	if exists {
		return false
	} else {
		//search-index don't exist
		_, err := es.CreateIndex(index).BodyString(mapping).Do(ctx)
		log.GetTextLogger().Error("error creating index ,error is :", err.Error())
		return true
	}
}

// checkIndex check if search-index exist, exist? continue: re-create
func checkIndex(es *elastic.Client, index string) (bool, error) {
	exists, err := es.IndexExists(index).Do(ctx)
	if err != nil {
		log.GetTextLogger().Error("error is:", err.Error())
		return false, nil
	}
	return exists, nil

}

func (es *esClient) AddDoc(index string, data any) (bool, error) {
	exist, err := checkIndex(es.es, index)
	if err != nil {
		log.GetTextLogger().Fatal("create index error,error is: %v", err.Error())
		return false, err
	}

	if !exist {
		log.GetTextLogger().Fatal("cannot find target index: %v", err.Error())
	}
	_, err = es.es.Index().Index(index).BodyJson(data).Do(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (es *esClient) Ping(esUrl string) (int, error) {
	_, code, err := es.es.Ping(esUrl).Do(context.Background())
	return code, err
}
