# ShardUO #
ShardUO is an experimental Ultima Online server emulator written in the Go
programming language.

## Current Working Features ##
* Account creation and login
* Global chat

## Next Milestone - Mining Vertical ##
The next milestone set for ShardUO is to implement the Mining skill vertical,
including all features required to mine ore and smelt it into ingots.

### Required Features ###
* ~~Map loading~~
* Map tile query
* ~~Stats~~
* ~~Running~~
* Picking up and dropping items on the ground (partially implemented)
* ~~Player persistance~~
* ~~Removing equipment~~
* ~~Equipping equipment~~
* Using equipment
* Mining skill
* Smelting
* Backpack and inventory
* Container handling
* Spacial chat with speaking, yelling, and whispering

### Nice to Haves ###
* Gold
* NPC Vendors
* Stable Master NPC
* Smelter NPC
* Miner NPC
* Miner Guildmaster NPC
* Miner's Guild
* Buying from vendors
* Mounts
* Horses and Llamas
* Pack Horses and Llamas
* Banker NPC
* Bank box
* Line of sight checks

## Outstanding Issues ##
This is a list of known issues that do not need to be resolved before the next
milestone but do need to be resolved sometime.

* Proper coordinate wrapping in game.Map.getChunksInBounds()
* Proper coordinate wrapping in uo.Location.XYDistance()
