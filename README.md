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
2. Create a `multi_sim_<timestamp>` directory
3. Run N independent simulations **in parallel** with the same scenario
4. Save each simulation to `simulation_N/` subdirectory with:
   - `actors.json` - Generated actors
   - `turn_N/actions.json` - Actions from each turn
   - `turn_N/world_state.json` - World state after each turn
   - `result.json` - Final result with yes/no answer
   - `simulation.log` - Full detailed log of the simulation
5. Save aggregate results to `aggregate_results.json`
6. Display aggregate statistics (percentage of yes/no outcomes)
7. Display one-paragraph summaries of each simulation result

Directory structure:
```
multi_sim_<timestamp>/
├── scenario.json              # Scenario, question, and turns
├── simulation_1/
│   ├── actors.json
│   ├── turn_1/
│   │   ├── actions.json
│   │   └── world_state.json
│   ├── turn_2/
│   │   └── ...
│   ├── result.json
│   └── simulation.log
├── simulation_2/
│   └── ...
└── aggregate_results.json
```

**Note:** In this mode, you cannot edit actors or actions between runs. This mode is designed for statistical analysis, not interactive editing.

**Performance:** Simulations run in parallel, significantly reducing total runtime. For example, 10 simulations run in roughly the time it takes to complete the slowest simulation, rather than 10x the time of a single simulation.

**Logging:** Each simulation writes its detailed output (actors, world state, actions) to its own `simulation.log` file. This allows parallel execution without log interleaving. Progress updates ("Starting simulation N", "Completed simulation N") are shown on the console.

**Output:** After all simulations complete, aggregate statistics and one-paragraph summaries of each simulation are displayed.

### Verbose Mode

Enable detailed logging for debugging:

```bash
./who-does-what --verbose
./who-does-what --num-simulations 10 --verbose
./who-does-what --interactive --verbose
```

The `--verbose` flag enables detailed logging of:
- HTTP requests to OpenAI API
- Schema generation
- JSON marshalling/unmarshalling
- Internal function calls

By default, verbose logging is disabled and only meaningful output (actor actions, world state, results) is shown.

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
- **Parallel Processing**:
  - Actor decisions within each turn are processed in parallel
  - Multiple simulations run in parallel for faster statistical analysis
- **LLM-Driven**: All major decisions powered by OpenAI's API
- **Robust API Calls**: Automatic retry with exponential backoff (up to 5 attempts) for failed requests
- **Interactive Editing**: Review and modify actors and turn data in real-time
- **Statistical Analysis**: Run multiple simulations to understand outcome distributions
- **Comprehensive Logging**: Each simulation gets its own detailed log file

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
# View one-paragraph summaries of each simulation result
```

Example output:
```
=== AGGREGATE RESULTS ===
Question: Did the Bank of Japan raise rates?
Total simulations: 10
Yes count: 7
No count: 3
Yes percentage: 70.0%

Results saved to: multi_sim_20260203_143022

=== INDIVIDUAL SIMULATION SUMMARIES ===

Simulation 1: Yes - The Bank of Japan raised interest rates by 0.25% after careful consideration of domestic economic stability and international pressures...
Simulation 2: No - The Bank of Japan maintained its current policy stance, citing concerns about economic recovery and the potential impact on the carry trade...
...
```

### Verbose logging for debugging
```bash
./who-does-what --verbose
./who-does-what --num-simulations 10 --verbose
# See detailed HTTP requests, schema generation, and internal operations
```

## Configuration

Set your OpenAI API key in a `.env` file:

```
OPENAI_API_KEY=your_api_key_here
```
