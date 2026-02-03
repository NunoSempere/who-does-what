# Who Does What

A scenario simulation tool that creates actors with goals and powers, simulates their interactions over multiple turns, and answers questions about the outcomes.

## Installation

```bash
go build -o who-does-what
```

## Usage

### Default Mode (Single Simulation)

Run a single simulation with the default scenario:

```bash
./who-does-what
```

### Interactive Mode

Run an interactive session where you can:
- Define your own scenario
- Edit actor files before simulation
- Review and edit each turn's data
- Save all data to a session folder

```bash
./who-does-what --interactive
```

The interactive mode will:
1. Ask for a scenario description
2. Ask for the number of turns
3. Ask for a summarization question
4. Create a `session_<pid>` directory
5. Generate actors and save them to `actors/` directory
6. Allow you to edit actor files
7. Run each turn and save data to `turn_N/` directories
8. Save the final result to `final_result.json`

### Multiple Simulations Mode

Run multiple simulations to get aggregate statistics:

```bash
./who-does-what --num-simulations 10
```

This will:
1. Prompt for scenario description, number of turns, and question
2. Run N independent simulations with the same scenario
3. Collect yes/no answers to the summarization question
4. Display aggregate statistics (percentage of yes/no outcomes)

**Note:** In this mode, actors and actions are NOT saved to files and you cannot edit them between runs. This mode is designed for statistical analysis, not interactive editing.

## Important Notes

**The `--interactive` and `--num-simulations` flags cannot be used together.** They represent different workflows:
- `--interactive`: Single simulation with full file editing capabilities
- `--num-simulations`: Multiple automated simulations for statistical analysis

Using both flags will result in an error.

## Architecture

The simulation follows these steps:

1. **Generate Actors**: Create actors with names, goals, and powers based on the scenario
2. **Initialize World State**: Create an initial state with events and description
3. **Simulate Turns**: For each turn:
   - Each actor observes their filtered view of the world (based on what they would know)
   - Each actor decides on an action based on their view
   - All actions are applied to update the world state
4. **Answer Question**: Analyze the final state to answer a yes/no question about the outcome

## Key Features

- **Information Asymmetry**: Actors only see what they would realistically know
- **Parallel Processing**: Actor decisions within each turn are processed in parallel
- **LLM-Driven**: All major decisions powered by OpenAI's API
- **Interactive Editing**: Review and modify actors and turn data in real-time
- **Statistical Analysis**: Run multiple simulations to understand outcome distributions

## Examples

### Running with default scenario
```bash
./who-does-what
```

### Interactive session
```bash
./who-does-what --interactive
# Follow the prompts to define your scenario
# Edit actor files and review turn data
```

### Statistical analysis
```bash
./who-does-what --num-simulations 100
# Enter your scenario, turns, and question when prompted
# Get percentage breakdown of outcomes across 100 automated simulations
```

## Configuration

Set your OpenAI API key in a `.env` file:

```
OPENAI_API_KEY=your_api_key_here
```
