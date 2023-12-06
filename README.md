# ShardUO #
ShardUO is an experimental Ultima Online server emulator written in the Go
programming language.

## Thanks and Recognition ##
A huge thank you goes out to the following groups, in no particular order.

* The Penultima Online (POL) packet documentation for being a guide stone and
  first source of packet and GUMP information.
* Every contributor to the RunUO / PlayUO / ServUO source code for packet
  reference.
* Every contributor to the ClassicUO source code for packet and client behavior
  reference.

## Current Working Features ##
* Account creation and login
* Global chat
* Spatial chat
* Walking, running and mounted movement
* Items and containers
* Equipment
* Generic GUMPs
* Mining
* Vendors
* Spawners
* AI models
* Doors, signs and decoration

## Next Milestone - Mining Vertical ##
The next milestone set for ShardUO is to implement the Mining skill vertical
including all features required to mine ore and smelt it into ingots. Many other
basic supporting features are also slated before the first alpha including
vendors, mounts, pack animals, and banking.

### Known Defects ###
* Server panics when trying to delete a character, it should just refuse

### Required Features and Tasks ###
* Panics are not being logged

### Nice to Haves ###
* Feedback GUMP
* Help GUMP
* Miner NPC
* Miner Guildmaster NPC
* Miner's Guild
* Wild flora
* Apple, pear, and peach trees
* Complete decorating Britain

### Nerdy Things I Might Do for Fun ###
* Player movement packet throttling
* Day/Night cycles
* Light sources
* Weather patterns
* Fall damage

### Miscellaneous Issues ###
This is a list of known issues that do not need to be resolved before the next
milestone but do need to be resolved sometime.

* Reduce allocations for map query operations
* Reduce or remove golang map usage in performance-critical areas
* Proper coordinate wrapping in uo.Location.XYDistance()
* Wrapped overworld

### Intentional Differences ###
Below are differences between how things worked in the UO:R era on live and how
things work with ShardUO.

* Vendors
  * do not resell what they buy from players
  * do not use vendor economy / supply and demand mechanics
  * always stock 999 of every item every time you buy from them
  * have unlimited inventory
  * will buy anything - this may change in the future
* Mining
  * automatically keeps mining until the vein is empty

## Feature Packages ##
Below are feature verticals that I want to implement next in priority order.

* Banking and Trade
  * Checks
  * Pay vendors from gold in bank
  * Secure trade
* Tailoring
  * Sheep pen in W brit
  * Functional spinning wheels supporting
    * Flax to thread
    * Cotton to thread
    * Wool to yarn
  * Functional looms
  * Shop NPCs
    * Weaver
    * Tailor
  * Crafting GUMP
  * Crafting system
    * Resource consumption function
    * Item creation
    * Normal, exceptional, and inferior quality
  * Cloth item crafting
  * Leather item crafting
* Blacksmithing
  * Blacksmithing item crafting
  * Shop NPCs
    * Blacksmith
    * Armorer
    * Weaponsmith
* Tinkering
  * Tinkered item crafting
  * Tinkerer NPC
* Lumberjacking
  * Wood resources
  * Lumberjacking skill
* Carpentry
  * Carpentry item crafting
  * Carpenter NPC

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
the public under the same license terms. This requirement to share the source
code goes into effect when the software is made available for network
connections regardless if a binary was distributed or not. I chose this
license because I want to share knowledge and appreciation of the Ultima Online
service during the late 1990's and early 2000's. I welcome anyone who would use
this repository or its associated services who share that appreciation.

That said below are some things I have learned about administrating this shard.

* On first run the server creates default configuration files. Edit them.
* It takes about a minute after first shard start for all of the ore to spawn.
