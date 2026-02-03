package main

import (
	"log"
	"github.com/joho/godotenv"
	"os"
	"fmt"
	"encoding/json"
	"sync"
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

type WorldState struct {
	Events []string `json:"events"`
	Description string `json:"description"`
}

type ActorView struct {
	VisibleEvents []string `json:"visible_events"`
	Interpretation string `json:"interpretation"`
}

type ActorAction struct {
	ActorName string `json:"actor_name"`
	Action string `json:"action"`
	Reasoning string `json:"reasoning"`
}

type SummarizationAnswer struct {
	Answer string `json:"answer"`
	YesNo bool `json:"yes_no"`
}

func GetActors(situation_description string, client *openai.Client) (Actors, error){
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
	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5, client: client}, openai_schema)
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

// AdjustActors takes existing actors and adjusts them based on external information
func AdjustActors(actors Actors, external_info string, client *openai.Client) (Actors, error) {
	log.Printf("[AdjustActors] Adjusting actors based on external information")

	actorsJSON, err := json.Marshal(actors)
	if err != nil {
		return Actors{}, fmt.Errorf("failed to marshal actors: %v", err)
	}

	prompt := fmt.Sprintf(`Given these actors: %s

And this new external information: %s

Please adjust the actors (their goals, powers, or add/remove actors) based on this new information. Return the adjusted list in the same JSON format.`, string(actorsJSON), external_info)

	var adjustedActors Actors
	schema, err := jsonschema.GenerateSchemaForType(adjustedActors)
	if err != nil {
		return Actors{}, fmt.Errorf("schema generation failed: %v", err)
	}

	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "Actors",
		Schema: schema,
		Strict: true,
	}

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5, client: client}, openai_schema)
	if err != nil {
		return Actors{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &adjustedActors)
	if err != nil {
		return Actors{}, err
	}

	log.Printf("[AdjustActors] Actors adjusted successfully")
	return adjustedActors, nil
}

// SummarizeWorldState creates a comprehensive summary of the current state of the world
func SummarizeWorldState(situation_description string, actors Actors, client *openai.Client) (WorldState, error) {
	log.Printf("[SummarizeWorldState] Creating world state summary")

	actorsJSON, err := json.Marshal(actors)
	if err != nil {
		return WorldState{}, fmt.Errorf("failed to marshal actors: %v", err)
	}

	prompt := fmt.Sprintf(`Given this situation: %s

And these actors: %s

Create a comprehensive summary of the current state of the world as a JSON object with:
- events: an array of specific events and facts about the current situation
- description: a general description of the overall state

Format: {"events": ["event 1", "event 2", ...], "description": "overall description"}`, situation_description, string(actorsJSON))

	var worldState WorldState
	schema, err := jsonschema.GenerateSchemaForType(worldState)
	if err != nil {
		return WorldState{}, fmt.Errorf("schema generation failed: %v", err)
	}

	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "WorldState",
		Schema: schema,
		Strict: true,
	}

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5, client: client}, openai_schema)
	if err != nil {
		return WorldState{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &worldState)
	if err != nil {
		return WorldState{}, err
	}

	log.Printf("[SummarizeWorldState] World state summarized successfully")
	return worldState, nil
}

// FilterWorldStateForActor takes the world state and an actor, and returns only the information
// that the actor would realistically know based on their position and powers
func FilterWorldStateForActor(worldState WorldState, actor Actor, client *openai.Client) (ActorView, error) {
	log.Printf("[FilterWorldStateForActor] Filtering world state for actor: %s", actor.Name)

	worldStateJSON, err := json.Marshal(worldState)
	if err != nil {
		return ActorView{}, fmt.Errorf("failed to marshal world state: %v", err)
	}

	actorJSON, err := json.Marshal(actor)
	if err != nil {
		return ActorView{}, fmt.Errorf("failed to marshal actor: %v", err)
	}

	prompt := fmt.Sprintf(`Given this complete world state: %s

And this actor: %s

Determine what information this actor would realistically know, see, or have access to based on their position and powers. Return a JSON object with:
- visible_events: array of events/information the actor would know about
- interpretation: how the actor interprets and understands the visible information given their goals

Only include information the actor would actually have access to. Some events might be completely unknown to them.`, string(worldStateJSON), string(actorJSON))

	var actorView ActorView
	schema, err := jsonschema.GenerateSchemaForType(actorView)
	if err != nil {
		return ActorView{}, fmt.Errorf("schema generation failed: %v", err)
	}

	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "ActorView",
		Schema: schema,
		Strict: true,
	}

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5, client: client}, openai_schema)
	if err != nil {
		return ActorView{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &actorView)
	if err != nil {
		return ActorView{}, err
	}

	log.Printf("[FilterWorldStateForActor] World state filtered successfully for %s", actor.Name)
	return actorView, nil
}

// ActorTakesAction has the actor decide what action to take based on their view of the world
func ActorTakesAction(actor Actor, actorView ActorView, client *openai.Client) (ActorAction, error) {
	log.Printf("[ActorTakesAction] Getting action for actor: %s", actor.Name)

	actorJSON, err := json.Marshal(actor)
	if err != nil {
		return ActorAction{}, fmt.Errorf("failed to marshal actor: %v", err)
	}

	actorViewJSON, err := json.Marshal(actorView)
	if err != nil {
		return ActorAction{}, fmt.Errorf("failed to marshal actor view: %v", err)
	}

	prompt := fmt.Sprintf(`Given this actor: %s

And their view of the world: %s

What action would this actor take given their goals, powers, and what they know? Return a JSON object with:
- actor_name: the name of the actor
- action: a description of the action they take
- reasoning: why they are taking this action given their goals and what they know`, string(actorJSON), string(actorViewJSON))

	var actorAction ActorAction
	schema, err := jsonschema.GenerateSchemaForType(actorAction)
	if err != nil {
		return ActorAction{}, fmt.Errorf("schema generation failed: %v", err)
	}

	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "ActorAction",
		Schema: schema,
		Strict: true,
	}

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5, client: client}, openai_schema)
	if err != nil {
		return ActorAction{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &actorAction)
	if err != nil {
		return ActorAction{}, err
	}

	log.Printf("[ActorTakesAction] Action determined for %s", actor.Name)

	// Print the action taken
	fmt.Printf("\n%s takes action: %s\n", actorAction.ActorName, actorAction.Action)
	fmt.Printf("Reasoning: %s\n", actorAction.Reasoning)

	return actorAction, nil
}

// RunSimulationTurn runs one turn of the simulation where each actor observes and acts
func RunSimulationTurn(worldState WorldState, actors Actors, client *openai.Client) ([]ActorAction, WorldState, error) {
	log.Printf("[RunSimulationTurn] Starting simulation turn with %d actors", len(actors.Actors))

	// Process each actor in parallel
	type actorResult struct {
		action ActorAction
		err    error
		index  int
	}

	results := make(chan actorResult, len(actors.Actors))
	var wg sync.WaitGroup

	// Each actor observes and acts in parallel
	for i, actor := range actors.Actors {
		wg.Add(1)
		go func(idx int, act Actor) {
			defer wg.Done()

			// Filter world state for this actor
			actorView, err := FilterWorldStateForActor(worldState, act, client)
			if err != nil {
				results <- actorResult{
					err:   fmt.Errorf("failed to filter world state for %s: %v", act.Name, err),
					index: idx,
				}
				return
			}

			// Actor takes action based on their view
			action, err := ActorTakesAction(act, actorView, client)
			if err != nil {
				results <- actorResult{
					err:   fmt.Errorf("failed to get action for %s: %v", act.Name, err),
					index: idx,
				}
				return
			}

			results <- actorResult{
				action: action,
				index:  idx,
			}
		}(i, actor)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	actions := make([]ActorAction, len(actors.Actors))
	for result := range results {
		if result.err != nil {
			return nil, worldState, result.err
		}
		actions[result.index] = result.action
	}

	// Update world state based on actions
	updatedWorldState, err := UpdateWorldState(worldState, actions, client)
	if err != nil {
		return actions, worldState, fmt.Errorf("failed to update world state: %v", err)
	}

	log.Printf("[RunSimulationTurn] Simulation turn completed")
	return actions, updatedWorldState, nil
}

// UpdateWorldState updates the world state based on the actions taken by actors
func UpdateWorldState(worldState WorldState, actions []ActorAction, client *openai.Client) (WorldState, error) {
	log.Printf("[UpdateWorldState] Updating world state based on %d actions", len(actions))

	worldStateJSON, err := json.Marshal(worldState)
	if err != nil {
		return WorldState{}, fmt.Errorf("failed to marshal world state: %v", err)
	}

	actionsJSON, err := json.Marshal(actions)
	if err != nil {
		return WorldState{}, fmt.Errorf("failed to marshal actions: %v", err)
	}

	prompt := fmt.Sprintf(`Given this world state: %s

And these actions taken by actors: %s

Update the world state to reflect the consequences of these actions. Return the updated world state in the same JSON format with:
- events: updated array of events including the consequences of the actions
- description: updated description of the overall state`, string(worldStateJSON), string(actionsJSON))

	var updatedWorldState WorldState
	schema, err := jsonschema.GenerateSchemaForType(updatedWorldState)
	if err != nil {
		return WorldState{}, fmt.Errorf("schema generation failed: %v", err)
	}

	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "WorldState",
		Schema: schema,
		Strict: true,
	}

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5, client: client}, openai_schema)
	if err != nil {
		return WorldState{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &updatedWorldState)
	if err != nil {
		return WorldState{}, err
	}

	log.Printf("[UpdateWorldState] World state updated successfully")
	return updatedWorldState, nil
}

// AnswerSummarizationQuestion answers a specific question about the final state of the simulation
func AnswerSummarizationQuestion(question string, worldState WorldState, allActions [][]ActorAction, client *openai.Client) (string, bool, error) {
	log.Printf("[AnswerSummarizationQuestion] Answering question: %s", question)

	worldStateJSON, err := json.Marshal(worldState)
	if err != nil {
		return "", false, fmt.Errorf("failed to marshal world state: %v", err)
	}

	allActionsJSON, err := json.Marshal(allActions)
	if err != nil {
		return "", false, fmt.Errorf("failed to marshal all actions: %v", err)
	}

	prompt := fmt.Sprintf(`Given this final world state: %s

And this history of all actions taken across turns: %s

Please answer this question: %s

Provide a JSON response with:
- answer: a clear detailed answer, referencing specific events and actions from the simulation
- yes_no: a boolean (true/false) indicating the yes/no answer to the question`, string(worldStateJSON), string(allActionsJSON), question)

	var summarizationAnswer SummarizationAnswer
	schema, err := jsonschema.GenerateSchemaForType(summarizationAnswer)
	if err != nil {
		return "", false, fmt.Errorf("schema generation failed: %v", err)
	}

	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "SummarizationAnswer",
		Schema: schema,
		Strict: true,
	}

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5, client: client}, openai_schema)
	if err != nil {
		return "", false, err
	}

	err = json.Unmarshal([]byte(openai_json), &summarizationAnswer)
	if err != nil {
		return "", false, err
	}

	log.Printf("[AnswerSummarizationQuestion] Question answered successfully")
	return summarizationAnswer.Answer, summarizationAnswer.YesNo, nil
}

func main(){

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
	openaiToken := os.Getenv("OPENAI_API_KEY")

	// Create OpenAI client once for reuse
	client := openai.NewClient(openaiToken)

	situation_description := "The Bank of Japan is considering what to do about rates. I am curious about how to balance the central bank of Japan changing rates with the needs of the Japanese people, the PM, but also possible external pressure to not unwind the Japanese carry trade."

	// Step 1: Get initial actors
	fmt.Println("\n=== STEP 1: Generating Actors ===")
	actors, err := GetActors(situation_description, client)
	if err != nil {
		log.Fatalf("Failed to get actors: %v", err)
	}
	pretty_actors, _ := json.MarshalIndent(actors, "", "  ")
	fmt.Printf("%v\n", string(pretty_actors))

	// Step 2: Optionally adjust actors based on external information (skipping for MVP demo)
	// external_info := "The US Federal Reserve has expressed concerns about global financial stability"
	// actors, _ = AdjustActors(actors, external_info, client)

	// Step 3: Summarize initial world state
	fmt.Println("\n=== STEP 2: Initial World State ===")
	worldState, err := SummarizeWorldState(situation_description, actors, client)
	if err != nil {
		log.Fatalf("Failed to summarize world state: %v", err)
	}
	pretty_world, _ := json.MarshalIndent(worldState, "", "  ")
	fmt.Printf("%v\n", string(pretty_world))

	// Step 4: Run simulation turns
	numTurns := 2
	var allActions [][]ActorAction

	for turn := 1; turn <= numTurns; turn++ {
		fmt.Printf("\n=== STEP 3.%d: Simulation Turn %d ===\n", turn+2, turn)

		actions, newWorldState, err := RunSimulationTurn(worldState, actors, client)
		if err != nil {
			log.Fatalf("Failed to run simulation turn %d: %v", turn, err)
		}

		worldState = newWorldState
		allActions = append(allActions, actions)

		fmt.Printf("\nActions taken in turn %d:\n", turn)
		for _, action := range actions {
			fmt.Printf("\n%s: %s\n", action.ActorName, action.Action)
			fmt.Printf("Reasoning: %s\n", action.Reasoning)
		}

		fmt.Printf("\nUpdated world state:\n")
		pretty_world, _ := json.MarshalIndent(worldState, "", "  ")
		fmt.Printf("%v\n", string(pretty_world))
	}

	// Step 5: Answer summarization question
	fmt.Println("\n=== STEP 5: Final Summarization ===")
	question := "Did the Bank of Japan raise rates, potentially unwinding the Japanese carry trade?"
	answer, yesNo, err := AnswerSummarizationQuestion(question, worldState, allActions, client)
	if err != nil {
		log.Fatalf("Failed to answer summarization question: %v", err)
	}
	fmt.Printf("\nQuestion: %s\n", question)
	fmt.Printf("Yes/No: %t\n", yesNo)
	fmt.Printf("Answer: %s\n", answer)
}

