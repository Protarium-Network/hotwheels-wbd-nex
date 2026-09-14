package globals

// NEX configuration for "Hot Wheels: World's Best Driver" (Wii U, Firebrand
// Games / Warner Bros. Interactive, 2013; internal project name
// "TeamHotWheels" - confirmed via embedded MSVC debug paths in Game.rpx).
//
// Unlike every other server in this project, no public NEX database
// (kinnay.github.io's nexwiiu.json, PretendoNetwork repos, GBATemp/wiiubrew)
// documents this title's access key or game server ID, and no literal
// value exists in the binary either (NEX is statically linked, same
// situation as every sibling title). This is first-party reverse
// engineering with no public starting point - see RECON.md.
const (
	// GameServerIDUSA/GameServerIDEUR are UNVERIFIED HYPOTHESES: the low 32
	// bits of each region's Title ID (USA 0005000010143300 / WUP-P-AHWE,
	// Europe 0005000010145100 / WUP-P-AHWP - the EUR release also covers
	// Australia; no Japan release exists), following the pattern that held
	// for this project's Art of Balance server. Several other multi-region
	// titles here (Xenoblade Chronicles X, Art of Balance itself) turned
	// out to need only ONE shared game server ID despite multiple regional
	// title IDs, so don't assume both entries are actually needed until
	// console testing confirms it region by region (see RECON.md).
	GameServerIDUSA = "10143300"
	GameServerIDEUR = "10145100"

	// AccessKey is UNKNOWN. Recovering it requires a bruteforce sweep
	// against a real captured PRUDPv1 packet (PretendoNetwork's
	// access-key-extractor -bruteforce, the same workflow that recovered
	// Art of Balance's 96900116) - see RECON.md. This placeholder will not
	// authenticate any console until replaced.
	AccessKey = "00000000"

	// NEX library version. UNKNOWN - no capture exists yet. Placeholder
	// starts at 3.4.0, matching this project's other 2013-era titles
	// (Trine 2, Sochi 2014) as the closest-era guess; expect to revise
	// after a real CONNECT/LoginEx capture, the same way every sibling
	// server pinned its real value.
	NEXMajor = 3
	NEXMinor = 4
	NEXPatch = 0
)
