package main

import (
	"log"
	"github.com/joho/godotenv"
	"os"
	"fmt"
	"encoding/json"
	openai "github.com/sashabaranov/go-openai"
	jsonschema "github.com/sashabaranov/go-openai/jsonschema"
)


type Actor struct {
	Name string `json:"name"`
	Goals string `json:"goals"`
	Powers string `json:"powers"`
}
type Actors struct {
	Actors []Actor
	Observations string `json:"observations"`
}

func GetActors(situation_description string, token string) (Actors, error){
	prompt := `Provide a list of the relevant actors and their goals as a JSON object \
	{
		actors: [
		{"name": "Name 1", "goals": "Description of goals", "powers": "Formal and informal powers"}
		{"name": "Name 2", "goals": "Description of goals 2", "powers": "Formal and informal powers"},
	  	...
		],
		observations: "any notes"
	}
	for the following situation: ` + situation_description

	var actors Actors
	schema, err := jsonschema.GenerateSchemaForType(actors)
	if err != nil {
		log.Printf("[GetActors] GenerateSchemaForType error: %v", err)
		return Actors{}, fmt.Errorf("schema generation failed: %v", err)
	}
	log.Printf("[GetActors] JSON schema generated successfully")
	openai_schema  := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "Actors",
		Schema: schema,
		Strict: true,
	}
	log.Printf("[GetActors] Making OpenAI API call with model: %s", GPT5)
	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5, token: token}, openai_schema)
	if err != nil {
		log.Printf("[GetActors] OpenAI API call failed: %v", err)
		return Actors{}, err
	}
	log.Printf("[GetActors] OpenAI API call successful - response length: %d chars", len(openai_json))
	
	log.Printf("[GetActors] Unmarshalling OpenAI response JSON")
	err = json.Unmarshal([]byte(openai_json), &actors)
	if err != nil {
		log.Printf("[GetActors] Error unmarshalling json: %v", err)
		log.Printf("[GetActors] Response JSON was: %v", openai_json)
		return Actors{}, err
	}
	log.Printf("[GetActors] JSON unmarshalled successfully")
	// log.Printf("[GetActors] Actors generation complete: %v", actors)
	return actors, nil
}

func main(){

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
  openaiToken := os.Getenv("OPENAI_API_KEY")

  situation_description := "The Bank of Japan is considering what to do about rates. I am curious about how to balance the central bank of Japan changing rates with the needs of the Japanese people, the PM, but also possible external pressure to not unwind the Japanese carry trade."
	if actors, err := GetActors(situation_description, openaiToken); err == nil {
		pretty_actors, err := json.MarshalIndent(actors, "", "  ")
		if err == nil {
			fmt.Printf("%v", string(pretty_actors))
		}
	}
}

