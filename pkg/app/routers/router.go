package routers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mozhuli/go-elastic/pkg/app/entity"
	"github.com/mozhuli/go-elastic/pkg/app/models"
	"github.com/mozhuli/go-elastic/pkg/app/utils"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

var Index = "search"

func GetHandler(w http.ResponseWriter, r *http.Request) {
	client := models.NewElasticClient(utils.ElasticUrl())
	dat := utils.BodyToJson(r)
	eType := dat["type"].(string)
	query_type := dat["query_type"].(string)
	child_type := dat["child_type"].(string)
	start_index := int(dat["start_index"].(float64))
	array_of_json := dat["query_json"].([]interface{})
	size := int(dat["size"].(float64))
	sorting, err1 := dat["sort"].(map[string]interface{})
	if err1 != true {
		panic(err1)
	}
	var fieldName string
	var sortType bool
	for i := range sorting {
		if i == "field" {
			fieldName = sorting[i].(string)
		} else if i == "asc" {
			sortType = true
		}
	}
	bq := elastic.NewBoolQuery()
	if query_type == "parent" {
		datRecord := array_of_json[0]
		res := datRecord.(map[string]interface{})
		key := res["key"].(string)
		value := res["value"].(string)

		matchChildQuery := elastic.NewHasChildQuery(child_type, elastic.NewMatchQuery(key, value)).
			InnerHit(elastic.NewInnerHit().Name("messages"))
		bq = bq.Must(elastic.NewMatchAllQuery())
		bq = bq.Filter(matchChildQuery)

	} else {
		//newQ := elastic.NewBoolQuery()
		for i := 0; i < len(array_of_json); i++ {
			datRecord := array_of_json[i]
			res := datRecord.(map[string]interface{})
			qType := res["query_type"].(string)
			matchQueryType := res["match"].(string)
			key := res["key"].(string)
			value := res["value"].(interface{})
			//switch vv := value.(type) {
			//case string:
			//
			//case int:
			//
			//case []interface{}:
			//	for i, u := range vv {
			//		fmt.Println(i, u)
			//	}
			//default:
			//	fmt.Println(k, "is of a type I don't know how to handle")
			//}
			var matchType *elastic.MatchQuery
			var termQuery *elastic.TermQuery
			var rangeQuery *elastic.RangeQuery
			match := 0
			switch matchQueryType {
			case "text":
				value := res["value"].(string)
				fmt.Println(value)
				matchType = elastic.NewMatchQuery(key, value)
				break
			case "keyword":
				match = 1
				termQuery = elastic.NewTermQuery(key, value)
				break
			case "range":
				match = 2
				rangeQuery = elastic.NewRangeQuery(key)
				valueRange := value.(map[string]interface{})

				for i := range valueRange {
					switch i {
					case "gte":
						rangeQuery = rangeQuery.Gte(valueRange[i])
						break
					case "gt":
						rangeQuery = rangeQuery.Gt(valueRange[i])
						break
					case "lte":
						rangeQuery = rangeQuery.Lte(valueRange[i])
						break
					case "lt":
						rangeQuery = rangeQuery.Lt(valueRange[i])
						break
					}

				}
				break
			}
			switch qType {
			case "must":
				if match == 0 {
					bq = bq.Must(matchType)
				} else {
					bq = bq.Must(termQuery)
				}
				break
			case "filter":
				if match == 0 {

					bq = bq.Filter(matchType)
				} else {
					bq = bq.Filter(termQuery)
				}
				break
			case "must_not":
				if match == 0 {

					bq = bq.MustNot(matchType)
				} else {
					bq = bq.MustNot(termQuery)

				}
				break
			case "should":
				if match == 0 {

					bq = bq.Should(matchType)
				} else {
					bq = bq.Should(termQuery)

				}
				break
			}
			//newQ = newQ.Should(matchType)
		}
		//bq.Filter(newQ)
	}

	fmt.Println(start_index, size)
	var searchResult *elastic.SearchResult
	eQuery := client.Search().
		Index(Index).
		Type(eType).
		Query(bq).From(start_index).
		Size(size)
	if fieldName != "" {
		eQuery = eQuery.Sort(fieldName, sortType)
	}
	searchResult, err := eQuery.Pretty(true).Do(context.Background())
	if err != nil {
		panic(err)
	}
	hits := searchResult.Hits.Hits

	datArray := make([]map[string]interface{}, len(hits))
	var dat1 map[string]interface{}

	for i := 0; i < len(hits); i++ {
		hit := searchResult.Hits.Hits[i]
		if err := json.Unmarshal(*hit.Source, &dat1); err != nil {
			panic(err)
		}
		fmt.Println(dat1)
		datArray[i] = dat1
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	response := entity.JsonResponse{"data_source": datArray, "status": true, "length": len(hits)}
	b, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(b))

}

func SetHandler(w http.ResponseWriter, r *http.Request) {
	dat := utils.BodyToJson(r)
	eType := dat["type"].(string)
	bodyData := dat["source"]
	id := dat["id"].(string)
	parent_id := dat["parent_id"].(string)
	operation := dat["operation"].(string)
	client := models.NewElasticClient(utils.ElasticUrl())
	indexService := client.Index().Index(Index)
	updateSevice := client.Update().Index(Index)
	deleteService := client.Delete().Index(Index)
	var err *error
	if operation == "add" {
		if parent_id != "" {
			indexService = indexService.Parent(parent_id)
		}
		_, _ = indexService.Id(id).Type(eType).BodyJson(bodyData).Do(context.Background())
	} else if operation == "update" {

		if parent_id != "" {
			updateSevice = updateSevice.Parent(parent_id)

		}

		_, _ = updateSevice.Type(eType).Id(id).Doc(bodyData).DetectNoop(true).Do(context.TODO())
	} else if operation == "delete" {
		if parent_id != "" {
			deleteService = deleteService.Id(id)
		}
		_, _ = deleteService.Type(eType).Do(context.TODO())
	}
	if err != nil {
		panic(err)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}

func MappingHandler(w http.ResponseWriter, r *http.Request) {
	client := models.NewElasticClient(utils.ElasticUrl())
	data := utils.BodyToJson(r)
	eType := data["entity"].(string)
	mappingData := data["mapping_json"].(map[string]interface{})
	//mappingIndex := elastic.NewAliasAddAction()
	putMappingResponse, err := client.PutMapping().Index(utils.DefaultIndex()).Type(eType).BodyJson(mappingData).Do(context.TODO())
	if err != nil {
		panic(err)
	}
	fmt.Println(putMappingResponse)
	newMapping, err := client.GetMapping().Index(utils.DefaultIndex()).Type(eType).Do(context.TODO())
	//var b map[string]interface{}
	b, marshalErr := json.Marshal(newMapping)
	if marshalErr != nil {
		panic(marshalErr)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	//w.Write([]byte("Gorilla Map!\n"))
	//fmt.Println(client.ClusterState())
	fmt.Fprint(w, string(b))

}
