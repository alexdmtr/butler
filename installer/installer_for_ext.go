package installer

// Note: .exe is not in that list because we need to
// read part of its contents to decide what we're going to
// do with it: 1) extract it 2) run it as an installer 3)
// just copy it naked
var installerForExt = map[string]InstallerType{

	///////////////////////////////////////////////////////////
	// Generic archives
	///////////////////////////////////////////////////////////

	".zip": InstallerTypeArchive,
	".gz":  InstallerTypeArchive,
	".bz2": InstallerTypeArchive,
	".7z":  InstallerTypeArchive,
	".tar": InstallerTypeArchive,
	".xz":  InstallerTypeArchive,
	".rar": InstallerTypeArchive,

	///////////////////////////////////////////////////////////
	// Known non-supported
	///////////////////////////////////////////////////////////

	".deb": InstallerTypeUnsupported,
	".rpm": InstallerTypeUnsupported,
	".pkg": InstallerTypeUnsupported,

	///////////////////////////////////////////////////////////
	// Platform-specific packages
	///////////////////////////////////////////////////////////

	// Apple disk images
	".dmg": InstallerTypeDMG,

	// Microsoft packages
	".msi": InstallerTypeMSI,

	///////////////////////////////////////////////////////////
	// Known naked that also sniff as other formats
	///////////////////////////////////////////////////////////

	".jar":          InstallerTypeNaked,
	".air":          InstallerTypeNaked,
	".love":         InstallerTypeNaked,
	".unitypackage": InstallerTypeNaked,

	///////////////////////////////////////////////////////////
	// Books!
	///////////////////////////////////////////////////////////

	".pdf":    InstallerTypeNaked,
	".ps":     InstallerTypeNaked,
	".djvu":   InstallerTypeNaked,
	".cbr":    InstallerTypeNaked,
	".cbz":    InstallerTypeNaked,
	".cb7":    InstallerTypeNaked,
	".cbt":    InstallerTypeNaked,
	".cba":    InstallerTypeNaked,
	".doc":    InstallerTypeNaked,
	".docx":   InstallerTypeNaked,
	".epub":   InstallerTypeNaked,
	".mobi":   InstallerTypeNaked,
	".pdb":    InstallerTypeNaked,
	".fb2":    InstallerTypeNaked,
	".xeb":    InstallerTypeNaked,
	".ceb":    InstallerTypeNaked,
	".ibooks": InstallerTypeNaked,
	".txt":    InstallerTypeNaked,

	///////////////////////////////////////////////////////////
	// Media
	///////////////////////////////////////////////////////////

	".ogg": InstallerTypeNaked,
	".mp3": InstallerTypeNaked,
	".wav": InstallerTypeNaked,
	".mp4": InstallerTypeNaked,
	".avi": InstallerTypeNaked,

	///////////////////////////////////////////////////////////
	// Images
	///////////////////////////////////////////////////////////

	".png": InstallerTypeNaked,
	".jpg": InstallerTypeNaked,
	".gif": InstallerTypeNaked,
	".bmp": InstallerTypeNaked,
	".tga": InstallerTypeNaked,

	///////////////////////////////////////////////////////////
	// Game Maker assets
	///////////////////////////////////////////////////////////

	".gmez": InstallerTypeNaked,
	".gmz":  InstallerTypeNaked,
	".yyz":  InstallerTypeNaked,
	".yymp": InstallerTypeNaked,

	///////////////////////////////////////////////////////////
	// ROMs
	///////////////////////////////////////////////////////////

	".gb":  InstallerTypeNaked,
	".gbc": InstallerTypeNaked,
	".sfc": InstallerTypeNaked,
	".smc": InstallerTypeNaked,
	".swc": InstallerTypeNaked,
	".gen": InstallerTypeNaked,
	".sg":  InstallerTypeNaked,
	".smd": InstallerTypeNaked,
	".md":  InstallerTypeNaked,

	///////////////////////////////////////////////////////////
	// Miscellaneous other things
	///////////////////////////////////////////////////////////

	// Some html games provide a single .html file
	// Now that's dedication.
	".html": InstallerTypeNaked,
}
