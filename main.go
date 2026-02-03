package main

import (
	"log"
	"github.com/joho/godotenv"
	"os"
	"fmt"
	"encoding/json"
	"sync"
	"flag"
	"bufio"
	"strings"
	"path/filepath"
	"io/ioutil"
	"time"
	openai "github.com/sashabaranov/go-openai"
	jsonschema "github.com/sashabaranov/go-openai/jsonschema"
)

var verbose bool

// Logger interface for simulation logging
type SimLogger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// FileLogger logs to a file
type FileLogger struct {
	logger *log.Logger
}

func NewFileLogger(file *os.File) *FileLogger {
	return &FileLogger{
		logger: log.New(file, "", 0),
	}
}

func (fl *FileLogger) Printf(format string, v ...interface{}) {
	fl.logger.Printf(format, v...)
}

func (fl *FileLogger) Println(v ...interface{}) {
	fl.logger.Println(v...)
}

// ConsoleLogger logs to stdout
type ConsoleLogger struct{}

func (cl *ConsoleLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (cl *ConsoleLogger) Println(v ...interface{}) {
	fmt.Println(v...)
}


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
		if verbose {
			log.Printf("[GetActors] GenerateSchemaForType error: %v", err)
		}
		return Actors{}, fmt.Errorf("schema generation failed: %v", err)
	}
	if verbose {
		log.Printf("[GetActors] JSON schema generated successfully")
	}
	openai_schema  := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "Actors",
		Schema: schema,
		Strict: true,
	}
	if verbose {
		log.Printf("[GetActors] Making OpenAI API call with model: %s", GPT5_2)
	}
	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_2, client: client}, openai_schema, verbose)
	if err != nil {
		if verbose {
			log.Printf("[GetActors] OpenAI API call failed: %v", err)
		}
		return Actors{}, err
	}
	if verbose {
		log.Printf("[GetActors] OpenAI API call successful - response length: %d chars", len(openai_json))
		log.Printf("[GetActors] Unmarshalling OpenAI response JSON")
	}

	err = json.Unmarshal([]byte(openai_json), &actors)
	if err != nil {
		if verbose {
			log.Printf("[GetActors] Error unmarshalling json: %v", err)
			log.Printf("[GetActors] Response JSON was: %v", openai_json)
		}
		return Actors{}, err
	}
	if verbose {
		log.Printf("[GetActors] JSON unmarshalled successfully")
	}
	return actors, nil
}

// AdjustActors takes existing actors and adjusts them based on external information
func AdjustActors(actors Actors, external_info string, client *openai.Client) (Actors, error) {
	if verbose {
		log.Printf("[AdjustActors] Adjusting actors based on external information")
	}

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

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_2, client: client}, openai_schema, verbose)
	if err != nil {
		return Actors{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &adjustedActors)
	if err != nil {
		return Actors{}, err
	}

	if verbose {
		log.Printf("[AdjustActors] Actors adjusted successfully")
	}
	return adjustedActors, nil
}

// SummarizeWorldState creates a comprehensive summary of the current state of the world
func SummarizeWorldState(situation_description string, actors Actors, client *openai.Client) (WorldState, error) {
	if verbose {
		log.Printf("[SummarizeWorldState] Creating world state summary")
	}

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

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_2, client: client}, openai_schema, verbose)
	if err != nil {
		return WorldState{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &worldState)
	if err != nil {
		return WorldState{}, err
	}

	if verbose {
		log.Printf("[SummarizeWorldState] World state summarized successfully")
	}
	return worldState, nil
}

// FilterWorldStateForActor takes the world state and an actor, and returns only the information
// that the actor would realistically know based on their position and powers
func FilterWorldStateForActor(worldState WorldState, actor Actor, client *openai.Client) (ActorView, error) {
	if verbose {
		log.Printf("[FilterWorldStateForActor] Filtering world state for actor: %s", actor.Name)
	}

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

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_2, client: client}, openai_schema, verbose)
	if err != nil {
		return ActorView{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &actorView)
	if err != nil {
		return ActorView{}, err
	}

	if verbose {
		log.Printf("[FilterWorldStateForActor] World state filtered successfully for %s", actor.Name)
	}
	return actorView, nil
}

// ActorTakesAction has the actor decide what action to take based on their view of the world
func ActorTakesAction(actor Actor, actorView ActorView, client *openai.Client, logger SimLogger) (ActorAction, error) {
	if verbose {
		log.Printf("[ActorTakesAction] Getting action for actor: %s", actor.Name)
	}

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

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_2, client: client}, openai_schema, verbose)
	if err != nil {
		return ActorAction{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &actorAction)
	if err != nil {
		return ActorAction{}, err
	}

	if verbose {
		log.Printf("[ActorTakesAction] Action determined for %s", actor.Name)
	}

	// Print the action taken to the logger
	if logger != nil {
		logger.Printf("\n%s takes action: %s\n", actorAction.ActorName, actorAction.Action)
		logger.Printf("Reasoning: %s\n", actorAction.Reasoning)
	}

	return actorAction, nil
}

// RunSimulationTurn runs one turn of the simulation where each actor observes and acts
func RunSimulationTurn(worldState WorldState, actors Actors, client *openai.Client, logger SimLogger) ([]ActorAction, WorldState, error) {
	if verbose {
		log.Printf("[RunSimulationTurn] Starting simulation turn with %d actors", len(actors.Actors))
	}

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
			action, err := ActorTakesAction(act, actorView, client, logger)
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

	if verbose {
		log.Printf("[RunSimulationTurn] Simulation turn completed")
	}
	return actions, updatedWorldState, nil
}

// UpdateWorldState updates the world state based on the actions taken by actors
func UpdateWorldState(worldState WorldState, actions []ActorAction, client *openai.Client) (WorldState, error) {
	if verbose {
		log.Printf("[UpdateWorldState] Updating world state based on %d actions", len(actions))
	}

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

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_2, client: client}, openai_schema, verbose)
	if err != nil {
		return WorldState{}, err
	}

	err = json.Unmarshal([]byte(openai_json), &updatedWorldState)
	if err != nil {
		return WorldState{}, err
	}

	if verbose {
		log.Printf("[UpdateWorldState] World state updated successfully")
	}
	return updatedWorldState, nil
}

// AnswerSummarizationQuestion answers a specific question about the final state of the simulation
func AnswerSummarizationQuestion(question string, worldState WorldState, allActions [][]ActorAction, client *openai.Client) (string, bool, error) {
	if verbose {
		log.Printf("[AnswerSummarizationQuestion] Answering question: %s", question)
	}

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

	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_2, client: client}, openai_schema, verbose)
	if err != nil {
		return "", false, err
	}

	err = json.Unmarshal([]byte(openai_json), &summarizationAnswer)
	if err != nil {
		return "", false, err
	}

	if verbose {
		log.Printf("[AnswerSummarizationQuestion] Question answered successfully")
	}
	return summarizationAnswer.Answer, summarizationAnswer.YesNo, nil
}

type SimulationResult struct {
	Question string
	YesNo    bool
	Answer   string
}

func runSingleSimulation(situationDescription string, question string, numTurns int, client *openai.Client, saveDir string, logger SimLogger) (SimulationResult, error) {
	// Use console logger if none provided
	if logger == nil {
		logger = &ConsoleLogger{}
	}

	// Step 1: Get initial actors
	logger.Println("\n=== Generating Actors ===")
	actors, err := GetActors(situationDescription, client)
	if err != nil {
		return SimulationResult{}, fmt.Errorf("failed to get actors: %v", err)
	}
	pretty_actors, _ := json.MarshalIndent(actors, "", "  ")
	logger.Printf("%v\n", string(pretty_actors))

	// Save actors if save directory is provided
	if saveDir != "" {
		actorsJSON, _ := json.MarshalIndent(actors, "", "  ")
		ioutil.WriteFile(filepath.Join(saveDir, "actors.json"), actorsJSON, 0644)
	}

	// Step 2: Summarize initial world state
	logger.Println("\n=== Initial World State ===")
	worldState, err := SummarizeWorldState(situationDescription, actors, client)
	if err != nil {
		return SimulationResult{}, fmt.Errorf("failed to summarize world state: %v", err)
	}
	pretty_world, _ := json.MarshalIndent(worldState, "", "  ")
	logger.Printf("%v\n", string(pretty_world))

	// Step 3: Run simulation turns
	var allActions [][]ActorAction

	for turn := 1; turn <= numTurns; turn++ {
		logger.Printf("\n=== Simulation Turn %d ===\n", turn)

		actions, newWorldState, err := RunSimulationTurn(worldState, actors, client, logger)
		if err != nil {
			return SimulationResult{}, fmt.Errorf("failed to run simulation turn %d: %v", turn, err)
		}

		worldState = newWorldState
		allActions = append(allActions, actions)

		logger.Printf("\nActions taken in turn %d:\n", turn)
		for _, action := range actions {
			logger.Printf("\n%s: %s\n", action.ActorName, action.Action)
			logger.Printf("Reasoning: %s\n", action.Reasoning)
		}

		logger.Printf("\nUpdated world state:\n")
		pretty_world, _ := json.MarshalIndent(worldState, "", "  ")
		logger.Printf("%v\n", string(pretty_world))

		// Save turn data if save directory is provided
		if saveDir != "" {
			turnDir := filepath.Join(saveDir, fmt.Sprintf("turn_%d", turn))
			os.MkdirAll(turnDir, 0755)

			actionsJSON, _ := json.MarshalIndent(actions, "", "  ")
			ioutil.WriteFile(filepath.Join(turnDir, "actions.json"), actionsJSON, 0644)

			worldStateJSON, _ := json.MarshalIndent(worldState, "", "  ")
			ioutil.WriteFile(filepath.Join(turnDir, "world_state.json"), worldStateJSON, 0644)
		}
	}

	// Step 4: Answer summarization question
	logger.Println("\n=== Final Summarization ===")
	answer, yesNo, err := AnswerSummarizationQuestion(question, worldState, allActions, client)
	if err != nil {
		return SimulationResult{}, fmt.Errorf("failed to answer summarization question: %v", err)
	}
	logger.Printf("\nQuestion: %s\n", question)
	logger.Printf("Yes/No: %t\n", yesNo)
	logger.Printf("Answer: %s\n", answer)

	result := SimulationResult{
		Question: question,
		YesNo:    yesNo,
		Answer:   answer,
	}

	// Save result if save directory is provided
	if saveDir != "" {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		ioutil.WriteFile(filepath.Join(saveDir, "result.json"), resultJSON, 0644)
	}

	return result, nil
}

func runInteractiveSimulation(client *openai.Client) error {
	reader := bufio.NewReader(os.Stdin)

	// Get scenario from user
	fmt.Print("\nEnter the scenario description: ")
	situationDescription, _ := reader.ReadString('\n')
	situationDescription = strings.TrimSpace(situationDescription)

	// Get number of turns
	fmt.Print("Enter number of turns to simulate: ")
	var numTurns int
	fmt.Scanf("%d\n", &numTurns)

	// Get summarization question
	fmt.Print("Enter the question to answer at the end: ")
	question, _ := reader.ReadString('\n')
	question = strings.TrimSpace(question)

	// Create session directory
	sessionDir := fmt.Sprintf("session_%d", os.Getpid())
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %v", err)
	}
	fmt.Printf("\nSession directory: %s\n", sessionDir)

	// Step 1: Get and save actors
	fmt.Println("\n=== Generating Actors ===")
	actors, err := GetActors(situationDescription, client)
	if err != nil {
		return fmt.Errorf("failed to get actors: %v", err)
	}

	// Save actors to files
	actorsDir := filepath.Join(sessionDir, "actors")
	if err := os.MkdirAll(actorsDir, 0755); err != nil {
		return fmt.Errorf("failed to create actors directory: %v", err)
	}

	for i, actor := range actors.Actors {
		actorJSON, _ := json.MarshalIndent(actor, "", "  ")
		actorFile := filepath.Join(actorsDir, fmt.Sprintf("actor_%d_%s.json", i+1, strings.ReplaceAll(actor.Name, " ", "_")))
		if err := ioutil.WriteFile(actorFile, actorJSON, 0644); err != nil {
			return fmt.Errorf("failed to write actor file: %v", err)
		}
	}

	fmt.Printf("\nActors saved to %s\n", actorsDir)
	fmt.Print("You can now edit the actor files. Press Enter when ready to continue...")
	reader.ReadString('\n')

	// Reload actors from files
	files, err := ioutil.ReadDir(actorsDir)
	if err != nil {
		return fmt.Errorf("failed to read actors directory: %v", err)
	}

	actors.Actors = []Actor{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			actorJSON, err := ioutil.ReadFile(filepath.Join(actorsDir, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to read actor file: %v", err)
			}
			var actor Actor
			if err := json.Unmarshal(actorJSON, &actor); err != nil {
				return fmt.Errorf("failed to unmarshal actor: %v", err)
			}
			actors.Actors = append(actors.Actors, actor)
		}
	}

	fmt.Printf("\nReloaded %d actors\n", len(actors.Actors))

	// Step 2: Summarize initial world state
	fmt.Println("\n=== Initial World State ===")
	worldState, err := SummarizeWorldState(situationDescription, actors, client)
	if err != nil {
		return fmt.Errorf("failed to summarize world state: %v", err)
	}

	// Step 3: Run simulation turns
	var allActions [][]ActorAction

	for turn := 1; turn <= numTurns; turn++ {
		turnDir := filepath.Join(sessionDir, fmt.Sprintf("turn_%d", turn))
		if err := os.MkdirAll(turnDir, 0755); err != nil {
			return fmt.Errorf("failed to create turn directory: %v", err)
		}

		fmt.Printf("\n=== Simulation Turn %d ===\n", turn)

		consoleLogger := &ConsoleLogger{}
		actions, newWorldState, err := RunSimulationTurn(worldState, actors, client, consoleLogger)
		if err != nil {
			return fmt.Errorf("failed to run simulation turn %d: %v", turn, err)
		}

		worldState = newWorldState
		allActions = append(allActions, actions)

		// Save turn data to files
		for i, action := range actions {
			actionJSON, _ := json.MarshalIndent(action, "", "  ")
			actionFile := filepath.Join(turnDir, fmt.Sprintf("action_%d_%s.json", i+1, strings.ReplaceAll(action.ActorName, " ", "_")))
			ioutil.WriteFile(actionFile, actionJSON, 0644)
		}

		worldStateJSON, _ := json.MarshalIndent(worldState, "", "  ")
		worldStateFile := filepath.Join(turnDir, "world_state.json")
		ioutil.WriteFile(worldStateFile, worldStateJSON, 0644)

		fmt.Printf("\nTurn %d data saved to %s\n", turn, turnDir)
		fmt.Print("Press Enter to continue to next turn...")
		reader.ReadString('\n')
	}

	// Step 4: Answer summarization question
	fmt.Println("\n=== Final Summarization ===")
	answer, yesNo, err := AnswerSummarizationQuestion(question, worldState, allActions, client)
	if err != nil {
		return fmt.Errorf("failed to answer summarization question: %v", err)
	}

	result := SimulationResult{
		Question: question,
		YesNo:    yesNo,
		Answer:   answer,
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	resultFile := filepath.Join(sessionDir, "final_result.json")
	ioutil.WriteFile(resultFile, resultJSON, 0644)

	fmt.Printf("\nQuestion: %s\n", question)
	fmt.Printf("Yes/No: %t\n", yesNo)
	fmt.Printf("Answer: %s\n", answer)
	fmt.Printf("\nFinal result saved to %s\n", resultFile)

	return nil
}

func runMultipleSimulations(numSimulations int, client *openai.Client) error {
	reader := bufio.NewReader(os.Stdin)

	// Get scenario from user
	fmt.Print("\nEnter the scenario description: ")
	situationDescription, _ := reader.ReadString('\n')
	situationDescription = strings.TrimSpace(situationDescription)

	// Get number of turns
	fmt.Print("Enter number of turns to simulate: ")
	var numTurns int
	fmt.Scanf("%d\n", &numTurns)

	// Get summarization question
	fmt.Print("Enter the question to answer at the end: ")
	question, _ := reader.ReadString('\n')
	question = strings.TrimSpace(question)

	// Create base directory for all simulations
	timestamp := time.Now().Format("20060102_150405")
	baseDir := fmt.Sprintf("multi_sim_%s", timestamp)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %v", err)
	}

	fmt.Printf("\n=== Running %d Simulations in Parallel ===\n", numSimulations)
	fmt.Printf("Scenario: %s\n", situationDescription)
	fmt.Printf("Turns: %d\n", numTurns)
	fmt.Printf("Question: %s\n", question)
	fmt.Printf("Saving to: %s\n", baseDir)

	// Save scenario information to base directory
	scenarioInfo := map[string]interface{}{
		"scenario": situationDescription,
		"question": question,
		"turns":    numTurns,
	}
	scenarioJSON, _ := json.MarshalIndent(scenarioInfo, "", "  ")
	ioutil.WriteFile(filepath.Join(baseDir, "scenario.json"), scenarioJSON, 0644)

	// Run simulations in parallel
	type simResult struct {
		index  int
		result SimulationResult
		err    error
	}

	resultsChan := make(chan simResult, numSimulations)
	var wg sync.WaitGroup

	for i := 0; i < numSimulations; i++ {
		wg.Add(1)
		go func(simIndex int) {
			defer wg.Done()

			fmt.Printf("Starting simulation %d/%d...\n", simIndex+1, numSimulations)

			// Create directory for this simulation
			simDir := filepath.Join(baseDir, fmt.Sprintf("simulation_%d", simIndex+1))
			if err := os.MkdirAll(simDir, 0755); err != nil {
				resultsChan <- simResult{
					index: simIndex,
					err:   fmt.Errorf("failed to create simulation directory: %v", err),
				}
				return
			}

			// Create log file for this simulation
			logFile, err := os.Create(filepath.Join(simDir, "simulation.log"))
			if err != nil {
				resultsChan <- simResult{
					index: simIndex,
					err:   fmt.Errorf("failed to create log file: %v", err),
				}
				return
			}
			defer logFile.Close()

			// Create file logger for this simulation
			logger := NewFileLogger(logFile)

			result, err := runSingleSimulation(situationDescription, question, numTurns, client, simDir, logger)
			if err != nil {
				resultsChan <- simResult{
					index: simIndex,
					err:   fmt.Errorf("simulation failed: %v", err),
				}
				return
			}

			fmt.Printf("Completed simulation %d/%d\n", simIndex+1, numSimulations)

			resultsChan <- simResult{
				index:  simIndex,
				result: result,
			}
		}(i)
	}

	// Wait for all simulations to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	results := make([]SimulationResult, numSimulations)
	yesCount := 0
	for res := range resultsChan {
		if res.err != nil {
			return fmt.Errorf("simulation %d failed: %v", res.index+1, res.err)
		}
		results[res.index] = res.result
		if res.result.YesNo {
			yesCount++
		}
	}

	// Aggregate results
	aggregateResult := map[string]interface{}{
		"question":           question,
		"total":              numSimulations,
		"yes_count":          yesCount,
		"no_count":           numSimulations - yesCount,
		"yes_percentage":     float64(yesCount) / float64(numSimulations) * 100,
		"scenario":           situationDescription,
		"turns":              numTurns,
		"individual_results": results,
	}

	aggregateJSON, _ := json.MarshalIndent(aggregateResult, "", "  ")
	ioutil.WriteFile(filepath.Join(baseDir, "aggregate_results.json"), aggregateJSON, 0644)

	fmt.Printf("\n\n=== AGGREGATE RESULTS ===\n")
	fmt.Printf("Question: %s\n", question)
	fmt.Printf("Total simulations: %d\n", numSimulations)
	fmt.Printf("Yes count: %d\n", yesCount)
	fmt.Printf("No count: %d\n", numSimulations-yesCount)
	fmt.Printf("Yes percentage: %.1f%%\n", float64(yesCount)/float64(numSimulations)*100)
	fmt.Printf("\nResults saved to: %s\n", baseDir)

	// Print one-paragraph summary for each simulation
	fmt.Printf("\n=== INDIVIDUAL SIMULATION SUMMARIES ===\n")
	for i, result := range results {
		yesNoStr := "No"
		if result.YesNo {
			yesNoStr = "Yes"
		}
		// Truncate answer to ~200 characters for summary
		summary := result.Answer
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}
		fmt.Printf("\nSimulation %d: %s - %s\n", i+1, yesNoStr, summary)
	}

	return nil
}

func main() {
	// Parse command-line flags
	interactive := flag.Bool("interactive", false, "Run in interactive mode")
	numSimulations := flag.Int("num-simulations", 0, "Run multiple simulations and aggregate results")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	verbose = *verboseFlag

	if err := godotenv.Load(".env"); err != nil {
		if verbose {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}
	openaiToken := os.Getenv("OPENAI_API_KEY")

	// Create OpenAI client once for reuse
	client := openai.NewClient(openaiToken)

	if *interactive && *numSimulations > 0 {
		log.Fatalf("Error: --interactive and --num-simulations flags cannot be used together")
	}

	if *interactive {
		// Run in interactive mode
		if err := runInteractiveSimulation(client); err != nil {
			log.Fatalf("Interactive simulation failed: %v", err)
		}
	} else if *numSimulations > 0 {
		// Run multiple simulations
		if err := runMultipleSimulations(*numSimulations, client); err != nil {
			log.Fatalf("Multiple simulations failed: %v", err)
		}
	} else {
		// Run single simulation with default scenario
		situationDescription := "The Bank of Japan is considering what to do about rates. I am curious about how to balance the central bank of Japan changing rates with the needs of the Japanese people, the PM, but also possible external pressure to not unwind the Japanese carry trade."
		question := "Did the Bank of Japan raise rates, potentially unwinding the Japanese carry trade?"
		numTurns := 2

		_, err := runSingleSimulation(situationDescription, question, numTurns, client, "", nil)
		if err != nil {
			log.Fatalf("Simulation failed: %v", err)
		}
	}
}

