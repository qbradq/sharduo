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
* Following the map during movement
* ~~Stats~~
* ~~Running~~
* ~~Picking up and dropping items on the ground~~
* ~~Player persistance~~
* ~~Removing equipment~~
* ~~Equipping equipment~~
* Action rate-limiting
* Using equipment
* Timers
* Mining skill
* Smelting
* ~~Backpack and inventory~~
* ~~Container handling~~
* Spacial chat with speaking, yelling, and whispering
* Mobile weight limit and stamina
* ~~Stat updates~~
* Skill updates
* Skill gains
* Stat gains
* ~~Stackable items~~
* Respond to status closed messages

### Nice to Haves ###
* ~~Gold~~
* NPC Vendors
* Stable Master NPC
* Smelter NPC
* Miner NPC
* Miner Guildmaster NPC
* Miner's Guild
* Buying from vendors
* ~~Mounts~~
* ~~Horses and Llamas~~
* Walk Random AI
* Follow Master AI
* ~~Pack Horses and Llamas~~
* Banker NPC
* ~~Bank box~~
* Line of sight checks
* Paper Dolls for others
* Item durability

## Outstanding Issues ##

### Go-Live Issues ###
This is a list of known issues that must be resolved before the server may be
opened to outside connections.

* Player movement packet throttling
* ~~Removal of stale NetStates in the game service goroutine~~
* ~~Removal of stale connections in the login service goroutine~~
* ~~Integration of the game service into a single self-managed structure~~
* Restrict command usage to access levels
* ~~Do not allow items within the bank box to be used unless the bank box is open~~
* ~~Proper coordinate wrapping in game.Map.getChunksInBounds()~~

### Miscellaneous Issues ###
This is a list of known issues that do not need to be resolved before the next
milestone but do need to be resolved sometime.

* Reduce redundancy in save files... somehow
* Improve client file load time
* Proper coordinate wrapping in uo.Location.XYDistance()
* Wrapped overworld

# Shard Administration Advice #
The ShardUO software and the creative content within were developed to host a
single service called Trammel Time. Using this codebase to create a different
service will require changes to configuration files, data files, and source
code at minimum. To create a service with different - read standard - rules will
also require extensive modification to the internal/game package, several
portions of the internal/cmd/uod package, and many constant updates in the
lib/uo package. To create a service that supports current Broadsword server
content would require nearly a complete re-write.

This repository is licensed under the AGPLv3 or later (see LICENSE). This
license requires that any modification to this source code be made available to
the public under the same license terms. This goes into effect the moment you
distribute a binary of the program in any form to anyone through any means or
you make a service based on this source code available for connection on any
network reachable by third-parties regardless if anyone connected. I chose this
license because I want to share knowledge and appreciation of the Ultima Online
service during the late 90's and early 20's. I welcome anyone who would use this
repository or its associated services who share that appreciation.

That said below are some things I have learned about administrating this shard.

