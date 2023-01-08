package uo

// SkillDefinitino holds all of the information about a skill
type SkillDefinition struct {
	// ID of the skill
	ID Skill
	// Name of the skill
	Name string
	// Title given by the skill
	Title string
	// Strength gain rate
	StrengthGain float32
	// Dexterity gain
	DexterityGain float32
	// Intelligence gain
	IntelligenceGain float32
	// Primary stat to gain
	PrimaryStat Stat
	// Secondary stat to gain
	SecondaryStat Stat
}

var SkillInfo = []*SkillDefinition{
	{SkillAlchemy, "Alchemy", "Alchemist", 0.0, 0.5, 0.5, StatIntelligence, StatDexterity},
	{SkillAnatomy, "Anatomy", "Biologist", 0.15, 0.15, 0.7, StatIntelligence, StatStrength},
	{SkillAnimalLore, "Animal Lore", "Naturalist", 0.0, 0.0, 1.0, StatIntelligence, StatStrength},
	{SkillItemID, "Item Identification", "Merchant", 0.0, 0.0, 1.0, StatIntelligence, StatDexterity},
	{SkillArmsLore, "Arms Lore", "Weapon Master", 0.75, 0.15, 0.1, StatIntelligence, StatStrength},
	{SkillParrying, "Parrying", "Duelist", 0.75, 0.25, 0.0, StatDexterity, StatStrength},
	{SkillBegging, "Begging", "Beggar", 0.0, 0.0, 0.0, StatDexterity, StatIntelligence},
	{SkillBlacksmithing, "Blacksmithy", "Blacksmith", 1.0, 0.0, 0.0, StatStrength, StatDexterity},
	{SkillBowcraft, "Bowcraft/Fletching", "Bowyer", 0.6, 1.6, 0.0, StatDexterity, StatStrength},
	{SkillPeacemaking, "Peacemaking", "Pacifier", 0.0, 0.0, 0.0, StatIntelligence, StatDexterity},
	{SkillCamping, "Camping", "Explorer", 2.0, 1.5, 1.5, StatDexterity, StatIntelligence},
	{SkillCarpentry, "Carpentry", "Carpenter", 2.0, 0.5, 0.0, StatStrength, StatDexterity},
	{SkillCartography, "Cartography", "Cartographer", 0.0, 0.75, 0.75, StatIntelligence, StatDexterity},
	{SkillCooking, "Cooking", "Chef", 0.0, 2.0, 3.0, StatIntelligence, StatDexterity},
	{SkillDetectHidden, "Detecting Hidden", "Scout", 0.0, 0.4, 0.6, StatIntelligence, StatDexterity},
	{SkillEnticement, "Discordance", "Demoralizer", 0.0, 0.25, 0.25, StatDexterity, StatIntelligence},
	{SkillEvaluateIntelligence, "Evaluating Intelligence", "Scholar", 0.0, 0.0, 1.0, StatIntelligence, StatStrength},
	{SkillHealing, "Healing", "Healer", 0.6, 0.6, 0.8, StatIntelligence, StatDexterity},
	{SkillFishing, "Fishing", "Fisherman", 0.5, 0.5, 0.0, StatDexterity, StatStrength},
	{SkillForensicEvaluation, "Forensic Evaluation", "Detective", 0.0, 0.2, 0.8, StatIntelligence, StatDexterity},
	{SkillHerding, "Herding", "Shepherd", 1.625, 0.625, 0.25, StatIntelligence, StatDexterity},
	{SkillHiding, "Hiding", "Shade", 0.0, 0.8, 0.2, StatDexterity, StatIntelligence},
	{SkillProvocation, "Provocation", "Rouser", 0.0, 0.45, 0.05, StatIntelligence, StatDexterity},
	{SkillInscription, "Inscription", "Scribe", 0.0, 0.2, 0.8, StatIntelligence, StatDexterity},
	{SkillLockpicking, "Lockpicking", "Infiltrator", 0.0, 2.0, 0.0, StatDexterity, StatIntelligence},
	{SkillMagery, "Magery", "Mage", 0.0, 0.0, 1.5, StatIntelligence, StatStrength},
	{SkillMagicResistance, "Resisting Spells", "Warder", 0.25, 0.25, 0.5, StatStrength, StatDexterity},
	{SkillTactics, "Tactics", "Tactician", 0.0, 0.0, 0.0, StatStrength, StatDexterity},
	{SkillSnooping, "Snooping", "Spy", 0.0, 2.5, 0.0, StatDexterity, StatIntelligence},
	{SkillMusicianship, "Musicianship", "Bard", 0.0, 0.8, 0.2, StatDexterity, StatIntelligence},
	{SkillPoisoning, "Poisoning", "Assassin", 0.0, 0.4, 1.6, StatIntelligence, StatDexterity},
	{SkillArchery, "Archery", "Archer", 0.25, 0.75, 0.0, StatDexterity, StatStrength},
	{SkillSpiritSpeak, "Spirit Speak", "Medium", 0.0, 0.0, 1.0, StatIntelligence, StatStrength},
	{SkillStealing, "Stealing", "Pickpocket", 0.0, 1.0, 0.0, StatDexterity, StatIntelligence},
	{SkillTailoring, "Tailoring", "Tailor", 0.38, 1.63, 0.5, StatDexterity, StatIntelligence},
	{SkillAnimalTaming, "Animal Taming", "Tamer", 1.4, 0.2, 0.4, StatStrength, StatIntelligence},
	{SkillTasteIdentification, "Taste Identification", "Praegustator", 0.2, 0.0, 0.8, StatIntelligence, StatStrength},
	{SkillTinkering, "Tinkering", "Tinker", 0.5, 0.2, 0.3, StatDexterity, StatIntelligence},
	{SkillTracking, "Tracking", "Ranger", 0.0, 1.25, 1.25, StatIntelligence, StatDexterity},
	{SkillVeterinary, "Veterinary", "Veterinarian", 0.8, 0.4, 0.8, StatIntelligence, StatDexterity},
	{SkillSwordsmanship, "Swordsmanship", "Swordsman", 0.75, 0.25, 0.0, StatStrength, StatDexterity},
	{SkillMaceFighting, "Mace Fighting", "Armsman", 0.9, 0.1, 0.0, StatStrength, StatDexterity},
	{SkillFencing, "Fencing", "Fencer", 0.45, 0.55, 0.0, StatDexterity, StatStrength},
	{SkillWrestling, "Wrestling", "Wrestler", 0.9, 0.1, 0.0, StatStrength, StatDexterity},
	{SkillLumberjacking, "Lumberjacking", "Lumberjack", 2.0, 0.0, 0.0, StatStrength, StatDexterity},
	{SkillMining, "Mining", "Miner", 2.0, 0.0, 0.0, StatStrength, StatDexterity},
	{SkillMeditation, "Meditation", "Stoic", 0.0, 0.0, 0.0, StatIntelligence, StatStrength},
	{SkillStealth, "Stealth", "Rogue", 0.0, 0.0, 0.0, StatDexterity, StatIntelligence},
	{SkillRemoveTrap, "Remove Trap", "Trap Specialist", 0.0, 0.0, 0.0, StatDexterity, StatIntelligence},
	{SkillNecromancy, "Necromancy", "Necromancer", 0.0, 0.0, 0.0, StatIntelligence, StatStrength},
	{SkillFocus, "Focus", "Driven", 0.0, 0.0, 0.0, StatDexterity, StatIntelligence},
	{SkillChivalry, "Chivalry", "Paladin", 0.0, 0.0, 0.0, StatStrength, StatIntelligence},
	{SkillBushido, "Bushido", "Samurai", 0.0, 0.0, 0.0, StatStrength, StatIntelligence},
	{SkillNinjitsu, "Ninjitsu", "Ninja", 0.0, 0.0, 0.0, StatDexterity, StatIntelligence},
	{SkillSpellweaving, "Spellweaving", "Arcanist", 0.0, 0.0, 0.0, StatIntelligence, StatStrength},
	{SkillMysticism, "Mysticism", "Mystic", 0.0, 0.0, 0.0, StatStrength, StatIntelligence},
	{SkillImbuing, "Imbuing", "Artificer", 0.0, 0.0, 0.0, StatIntelligence, StatStrength},
	{SkillThrowing, "Throwing", "Bladeweaver", 0.0, 0.0, 0.0, StatDexterity, StatStrength},
}

// SkillNames is a list of names of the skills in order
var SkillNames = []string{
	"Alchemy",
	"Anatomy",
	"Animal Lore",
	"Item ID",
	"Arms Lore",
	"Parrying",
	"Begging",
	"Blacksmithing",
	"Bowcraft",
	"Peacemaking",
	"Camping",
	"Carpentry",
	"Cartography",
	"Cooking",
	"Detect Hidden",
	"Enticement",
	"Evaluate Intelligence",
	"Healing",
	"Fishing",
	"Forensic Evaluation",
	"Herding",
	"Hiding",
	"Provocation",
	"Inscription",
	"Lockpicking",
	"Magery",
	"Magic Resistance",
	"Tactics",
	"Snooping",
	"Musicianship",
	"Poisoning",
	"Archery",
	"Spirit Speak",
	"Stealing",
	"Tailoring",
	"Animal Taming",
	"Taste Identification",
	"Tinkering",
	"Tracking",
	"Veterinary",
	"Swordsmanship",
	"Mace Fighting",
	"Fencing",
	"Wrestling",
	"Lumberjacking",
	"Mining",
	"Meditation",
	"Stealth",
	"Remove Trap",
	"Necromancy",
	"Focus",
	"Chivalry",
	"Bushido",
	"Ninjitsu",
	"Spellweaving",
	"Mysticism",
	"Imbuing",
	"Throwing",
}

// Skill represents a skill by numeric ID
type Skill byte

// All valid values for SkillID
const (
	SkillAlchemy              Skill = 0
	SkillAnatomy              Skill = 1
	SkillAnimalLore           Skill = 2
	SkillItemID               Skill = 3
	SkillArmsLore             Skill = 4
	SkillParrying             Skill = 5
	SkillBegging              Skill = 6
	SkillBlacksmithing        Skill = 7
	SkillBowcraft             Skill = 8
	SkillPeacemaking          Skill = 9
	SkillCamping              Skill = 10
	SkillCarpentry            Skill = 11
	SkillCartography          Skill = 12
	SkillCooking              Skill = 13
	SkillDetectHidden         Skill = 14
	SkillEnticement           Skill = 15
	SkillEvaluateIntelligence Skill = 16
	SkillHealing              Skill = 17
	SkillFishing              Skill = 18
	SkillForensicEvaluation   Skill = 19
	SkillHerding              Skill = 20
	SkillHiding               Skill = 21
	SkillProvocation          Skill = 22
	SkillInscription          Skill = 23
	SkillLockpicking          Skill = 24
	SkillMagery               Skill = 25
	SkillMagicResistance      Skill = 26
	SkillTactics              Skill = 27
	SkillSnooping             Skill = 28
	SkillMusicianship         Skill = 29
	SkillPoisoning            Skill = 30
	SkillArchery              Skill = 31
	SkillSpiritSpeak          Skill = 32
	SkillStealing             Skill = 33
	SkillTailoring            Skill = 34
	SkillAnimalTaming         Skill = 35
	SkillTasteIdentification  Skill = 36
	SkillTinkering            Skill = 37
	SkillTracking             Skill = 38
	SkillVeterinary           Skill = 39
	SkillSwordsmanship        Skill = 40
	SkillMaceFighting         Skill = 41
	SkillFencing              Skill = 42
	SkillWrestling            Skill = 43
	SkillLumberjacking        Skill = 44
	SkillMining               Skill = 45
	SkillMeditation           Skill = 46
	SkillStealth              Skill = 47
	SkillRemoveTrap           Skill = 48
	SkillNecromancy           Skill = 49
	SkillFocus                Skill = 50
	SkillChivalry             Skill = 51
	SkillBushido              Skill = 52
	SkillNinjitsu             Skill = 53
	SkillSpellweaving         Skill = 54
	SkillMysticism            Skill = 55
	SkillImbuing              Skill = 56
	SkillThrowing             Skill = 57
	SkillFirst                Skill = SkillAlchemy
	SkillLast                 Skill = SkillThrowing
	SkillAll                  Skill = 0xFF // Asks for all skills in a status request
)
