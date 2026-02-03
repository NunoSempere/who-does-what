I want to create a small minimalist tool for setting up scenarios. The tool should

- Come up with key actors, their names, goals and capabilities (done)
- Adjust them based on external information
- Summarize the state of the world
- Run a variable number of turns where each actor sees and focuses on some part of reality
  - Have an explicit function that takes the state of the world, the actor, and removes information the actor wouldn't know
- Ask a summarization question in the end (e.g.: does the Japanese carry trade unwind)
