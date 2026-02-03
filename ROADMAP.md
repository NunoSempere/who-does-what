## Step 1: MVP

I want to create a small minimalist tool for setting up scenarios. The tool should

- Come up with key actors, their names, goals and capabilities (done)
- Adjust them based on external information
- Summarize the state of the world
- Run a variable number of turns where each actor sees and focuses on some part of reality
  - Have an explicit function that takes the state of the world, the actor, and removes information the actor wouldn't know
- Ask a summarization question in the end (e.g.: does the Japanese carry trade unwind)

## Step 2: CLI 

- Now, instead of being a program with a preloaded scenario, have the cli ask about the scenario to the user
- Then, create a folder structure for each particular session
- Produce the list of actors, and save each of them to a file. Allow the user to browse and edit these files
- Reload the edited actors into memory
- Play out each round. Each round should be saved to a folder, where each actor has a file with their view of the world, and the actions they take. Additionally, there should be a file correspondind to the final summary after each round. The user should again be able to edit these files before proceeding
- Again there should be a summarization question at the end (e.g., "did the Japanese carry trade unwind?")

## Step 3: Many simulations

Eventually, we will want to repeat the above many times in order to get a view of events. We will want to 

- Summarize each simulation run
- Aggregate them using some summary statistic (now: in what % of worlds does the Japanese central bank raise rates, thus unwinding the Japanese carry trade?)

## Step 4: Website

Eventually we will want to create a website with similar capabilities.
