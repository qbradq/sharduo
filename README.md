# ShardUO #
ShardUO is an experimental Ultima Online server emulator written in the Go
programming language.

## Thanks and Recognition ##
A huge thank you goes out to the following groups, in no particular order.

* The Penultima Online (POL) packet documentation for being a guide stone and
  first source of packet information.
* Every contributor to the RunUO / PlayUO / ServUO source code for packet
  reference.
* Every contributor to the ClassicUO source code for packet and client behavior
  reference.

## Current Working Features ##
* Account creation and login
* Global chat
* Spatial chat
* Walking, running, mounted movement
* Items and containers
* Equipment
* Generic GUMPs

## Next Milestone - Mining Vertical ##
The next milestone set for ShardUO is to implement the Mining skill vertical,
including all features required to mine ore and smelt it into ingots.

### Current Defects ###
* Using the Help gem button causes the client to crash from the test GUMP

### Required Features and Tasks ###
* Player movement packet throttling
* Action packet throttling
* Mining skill
* Ore resource map
* Mobile weight limit and stamina
* ~~Do not leak child objects on Remove()~~
* ~~Automatically archive and clean old save files~~
* ~~Auto-saves~~
* ~~Make number of marshal goroutines configurable~~
* ~~Restrict command usage to access levels~~
* ~~Respond to status closed messages~~
* ~~Map loading~~
* ~~Map tile query~~
* ~~Following the map during movement~~
* ~~Stats~~
* ~~Running~~
* ~~Picking up and dropping items on the ground~~
* ~~Player persistance~~
* ~~Removing equipment~~
* ~~Equipping equipment~~
* ~~Using equipment~~
* ~~Timers~~
* ~~Smelting~~
* ~~Backpack and inventory~~
* ~~Container handling~~
* ~~Spacial chat with speaking, yelling, whispering, and emotes~~
* ~~Global chat~~
* ~~Stat updates~~
* ~~Skill system~~
* ~~Skill updates~~
* ~~Skill gains~~
* ~~Stat gains~~
* ~~Stackable items~~
* ~~Animations~~
* ~~Sounds~~
* ~~Music~~
* ~~Clilocs~~

### Nice to Haves ###
* NPC name generation
* NPC vendors
* Buying from vendors
* Stable Master NPC
* Walk Random AI
* Follow Master AI
* Smelter NPC
* Miner NPC
* Miner Guildmaster NPC
* Miner's Guild
* Selling to vendors
* Item durability
* Passive fauna
* Wild flora
* Agricultural crops
* Regions
* Region spawning
* Compressed GUMP support
* Welcome GUMP
* Feedback GUMP
* Help GUMP
* Decoration GUMP
* Start decorating
* MegaCliloc roll-over info
* Line of sight checks
* ~~GUMP support~~
* ~~Gold~~
* ~~Context menus~~
* ~~Paper Dolls for others~~
* ~~Pack Horses and Llamas~~
* ~~Banker NPC~~
* ~~Bank box~~
* ~~Mounts~~
* ~~Horses and Llamas~~

### Nerdy Things I Might Do for Fun ###
* Day/Night cycles
* Weather patterns
* Fall damage
* ~~Global light levels~~
* ~~Personal light levels~~

## Outstanding Issues ##

### Go-Live Tasks ###
This is a list of known tasks that must be completed before the server may be
opened to outside connections.

* Evaluate all TODO tags remaining in the source code, create tasks for them and assign them to milestones
* Create homepage site and service
* Make client available for download somehow
* ~~Removal of stale NetStates in the game service goroutine~~
* ~~Removal of stale connections in the login service goroutine~~
* ~~Integration of the game service into a single self-managed structure~~
* ~~Do not allow items within the bank box to be used unless the bank box is open~~
* ~~Proper coordinate wrapping in game.Map.getChunksInBounds()~~

### Miscellaneous Issues ###
This is a list of known issues that do not need to be resolved before the next
milestone but do need to be resolved sometime.

* Reduce allocations for map query operations
* Reduce or remove golang map usage in performance-critical areas
* Proper coordinate wrapping in uo.Location.XYDistance()
* Wrapped overworld
* ~~Save system re-write~~

## Feature Packages ##
Below are feature verticals that I want to implement next in priority order.

* New skill system
  * Skill use
  * Skill XP gains
  * Skills GUMP
  * Leveling up skills
  * Skill point reset
* Tinkering
  * GUMP support
  * Crafting GUMP
  * Crafting system
    * Resource consumption function
    * Item creation
    * Normal, exceptional, and inferior quality
  * Tinkered item crafting
* Blacksmithing
  * Blacksmithing item crafting
* Lumberjacking
  * Wood resources
  * Lumberjacking skill
* Carpentry
  * Carpentry item crafting

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

