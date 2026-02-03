## Step 1: MVP [x] COMPLETED

I want to create a small minimalist tool for setting up scenarios. The tool should

- [x] Come up with key actors, their names, goals and capabilities
- [x] Adjust them based on external information
- [x] Summarize the state of the world
- [x] Run a variable number of turns where each actor sees and focuses on some part of reality
  - [x] Have an explicit function that takes the state of the world, the actor, and removes information the actor wouldn't know
- [x] Ask a summarization question in the end (e.g.: does the Japanese carry trade unwind)

## Step 2: CLI [x] COMPLETED

- [x] Now, instead of being a program with a preloaded scenario, have the cli ask about the scenario to the user
- [x] Then, create a folder structure for each particular session
- [x] Produce the list of actors, and save each of them to a file. Allow the user to browse and edit these files
- [x] Reload the edited actors into memory
- [x] Play out each round. Each round should be saved to a folder, where each actor has a file with their view of the world, and the actions they take. Additionally, there should be a file correspondind to the final summary after each round. The user should again be able to edit these files before proceeding
- [x] Again there should be a summarization question at the end (e.g., "did the Japanese carry trade unwind?")

**Usage:** `./who-does-what --interactive`

## Step 3: Many simulations [x] COMPLETED

Eventually, we will want to repeat the above many times in order to get a view of events. We will want to

- [x] Summarize each simulation run
- [x] Aggregate them using some summary statistic (now: in what % of worlds does the Japanese central bank raise rates, thus unwinding the Japanese carry trade?)
- [x] Run simulations in parallel, writting simulation-specific logs to their own folder, but a one paragraph summary of each simulation result to stdout after the aggregate statistics
- [x] Retry failing OpenAI requests with an exponential backoff, and stop at 5 tries.
- [x] Save the questions to the multiple simulations folder as well.

**Usage:** `./who-does-what --num-simulations 10`

## Step 4: Website

Eventually we will want to create a website with similar capabilities.

## Notes to self

- Actors are biased towards taking action, even when they wouldn't necessarily have fast OODA loops
- Unclear meaning of time
- No way to incorporate external information into digest yet.
  - Use AskNews
  - Perplexity is a good start, could integrate it more
- Ask about more information at the beginning
- Need to use this a bit more in order to calibrate it
- Neat as a minimalist piece of software
- Would be good to be able to ask more than one question at the end
