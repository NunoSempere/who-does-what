# Architecture

## Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                        INITIALIZATION                            │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │   GetActors()         │
                    │   Create actors with  │
                    │   goals & powers      │
                    └───────────┬───────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │  AdjustActors()       │ (optional)
                    │  Refine based on      │
                    │  external info        │
                    └───────────┬───────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │ SummarizeWorldState() │
                    │ Create initial world  │
                    │ state with events     │
                    └───────────┬───────────┘
                                │
┌───────────────────────────────┼───────────────────────────────┐
│                               ▼                                │
│                   SIMULATION LOOP (N turns)                    │
│                                                                 │
│   For each turn:                                                │
│   ┌─────────────────────────────────────────────────────────┐ │
│   │  For each actor:                                         │ │
│   │                                                           │ │
│   │    ┌────────────────────────────────────┐               │ │
│   │    │ FilterWorldStateForActor()         │               │ │
│   │    │ - Takes full world state           │               │ │
│   │    │ - Returns only what actor knows    │               │ │
│   │    │   based on position & powers       │               │ │
│   │    └──────────────┬─────────────────────┘               │ │
│   │                   │                                       │ │
│   │                   ▼                                       │ │
│   │    ┌────────────────────────────────────┐               │ │
│   │    │ ActorTakesAction()                 │               │ │
│   │    │ - Actor decides action             │               │ │
│   │    │ - Based on their limited view      │               │ │
│   │    │ - Driven by goals                  │               │ │
│   │    └──────────────┬─────────────────────┘               │ │
│   │                   │                                       │ │
│   │                   └──────────────┐                        │ │
│   │                                  │                        │ │
│   └──────────────────────────────────┼────────────────────────┘ │
│                                      │                          │
│                                      ▼                          │
│                   ┌────────────────────────────────────┐        │
│                   │ UpdateWorldState()                 │        │
│                   │ - Apply consequences of actions    │        │
│                   │ - Generate new events              │        │
│                   │ - Update description               │        │
│                   └──────────────┬─────────────────────┘        │
│                                  │                              │
│                                  └──────────┐                   │
│                                             │                   │
└─────────────────────────────────────────────┼───────────────────┘
                                              │
                                              ▼
                              ┌───────────────────────────┐
                              │ AnswerSummarizationQuestion()│
                              │ - Analyze final state     │
                              │ - Review all actions      │
                              │ - Answer specific question│
                              └───────────────────────────┘
```

## Key Components

### Data Structures

- **Actor**: Represents a participant with name, goals, and powers
- **Actors**: Collection of actors with observations
- **WorldState**: Global state with events and description
- **ActorView**: Filtered view of world state for a specific actor
- **ActorAction**: Action taken by an actor with reasoning

### Core Functions

1. **GetActors()**: Generates initial actors based on situation description
2. **AdjustActors()**: Refines actors based on new information
3. **SummarizeWorldState()**: Creates comprehensive world state summary
4. **FilterWorldStateForActor()**: Filters information based on what actor would realistically know
5. **ActorTakesAction()**: Actor decides action based on their limited view
6. **RunSimulationTurn()**: Orchestrates one full turn of simulation
7. **UpdateWorldState()**: Updates world state based on actions taken
8. **AnswerSummarizationQuestion()**: Answers questions about final simulation state

## Execution Modes

### Default Mode
Runs a single simulation with a hardcoded scenario. Useful for quick testing and development.

```bash
./who-does-what
```

### Interactive Mode (`--interactive`)
Provides a full interactive experience:
- User defines the scenario, number of turns, and question
- Creates a session directory with all simulation data
- Actors are saved to individual JSON files for editing
- User can review and edit actors before simulation starts
- Each turn's data is saved to a separate directory
- Final result is saved to `final_result.json`

Directory structure:
```
session_<pid>/
├── actors/
│   ├── actor_1_<name>.json
│   ├── actor_2_<name>.json
│   └── ...
├── turn_1/
│   ├── action_1_<actor>.json
│   ├── action_2_<actor>.json
│   ├── ...
│   └── world_state.json
├── turn_2/
│   └── ...
└── final_result.json
```

### Multiple Simulations Mode (`--num-simulations N`)
Runs N independent simulations with the same scenario and aggregates results:
- User is prompted for scenario, number of turns, and question
- Each simulation runs independently with the same parameters
- Actors and actions are NOT saved to files (no editing capability)
- Yes/No answers are collected
- Aggregate statistics are displayed (percentage breakdown)
- Useful for understanding probability distributions of outcomes

Example output:
```
=== AGGREGATE RESULTS ===
Question: Did the Bank of Japan raise rates?
Total simulations: 100
Yes count: 73
No count: 27
Yes percentage: 73.0%
```

**Note:** The `--interactive` and `--num-simulations` flags are mutually exclusive and cannot be used together.

## Design Principles

### Information Asymmetry
Each actor only sees what they would realistically know based on their position and powers. This is enforced by `FilterWorldStateForActor()`, which is crucial for realistic simulations.

### Turn-Based Simulation
Actions are resolved in turns:
1. All actors observe simultaneously (their filtered view)
2. All actors decide actions simultaneously (based on their view)
3. All actions are applied to update the world state
4. Repeat for next turn

### Parallel Processing
Within each turn, all actor observations and action decisions are processed in parallel using goroutines. This significantly speeds up simulations, especially with many actors.

### LLM-Driven Decisions
All major decisions (actor behavior, world state updates, information filtering) are made by the LLM to ensure realistic and nuanced simulation behavior.

### Client Reuse
A single OpenAI client is created and reused across all API calls for efficiency.
