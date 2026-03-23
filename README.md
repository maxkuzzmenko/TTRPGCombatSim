# TTRPGCombatSim
A tabletop RPG combat simulator for testing and calibrating combat mechanics. Spin up a party, throw them at some enemies and watch how the encounters play out with real dice rolls.

# How to use

1. Build and run the Docker container
   `docker-compose up`

2. SSH into the running container
   `ssh -p 2222 sim@localhost`
   The default password is `dragons!!1!` (you can change it with the `SIM_PASSWORD` environment variable)

3. The simulator will prompt you to create a party by choosing classes and races for your characters

4. Then you'll fight enemies and the simulator runs through the entire encounter automatically, rolling dice and applying damage until one side wins or loses

# How it works

Each character has four stats: Strength, Agility, Intelligence and Charisma. The values get set based on your class choice. Fighters get high Strength, Mages get high Intelligence, Rogues get high Agility and Healers get high Intelligence too.

When a character attacks, the simulator rolls 2d6 (or d12 if you configure it that way) and compares the result against their attack stat plus modifiers. Critical successes and critical failures can swing the encounter.

Healers are like mages but they don't deal damage. Instead they use intelligence-based healing to restore HP to their teammates. They can heal a single target or the whole party at once.

Your characters gain XP from fights and level up to get stronger. Characters who die stay dead for that encounter.

Enemies have their own stats and difficulty modifiers to make them harder or easier to fight.

# Configuration

Adjust these constants at the top of `sim.go` to tune encounter difficulty and balance:

```
XPPerLevel        = 5     // XP needed to level up
HealerSingleAmt   = 4     // Single target heal amount
HealerAllAmt      = 1     // Party-wide heal amount
PlayerHealAmt     = 2     // Non-healer healing a teammate
UseD12            = false // Use d12 instead of 2d6 for checks
```

Rebuild the container after making changes with `docker-compose up --build`

# Building after changes

When you modify `sim.go` with your own mechanics or balance tweaks, rebuild the binary with:

`./build.sh`

This compiles your Go code and places the binary directly into the `shared/` folder. The SSH container mounts this folder and reads from it, so your new binary will be used automatically on the next login. No need to restart anything.

# Remote access

You can expose the simulator over the internet using ngrok or similar port forwarding services. This lets you and your friends run encounters together remotely.

To use ngrok:

`ngrok tcp 2222`

This gives you a remote hostname that forwards to your local SSH server. Others can SSH in with the hostname ngrok provides.

The SSH container has some security measures in place but is designed with local testing in mind. Use it at your own risk if exposing it over the internet. Change the default password and only share access with people you trust.

# What this is for

This tool lets you quickly test how encounters feel with different party compositions and enemy configurations. You can see how healing mechanics work, how different stat distributions affect combat length and difficulty, and whether your encounter balance needs tweaking. Much faster than running it through a real game session.
