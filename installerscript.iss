[Files]
Source: "release\huntstats.exe"; DestDir: "{app}"; Components: main
Source: "release\config.toml"; DestDir: "{app}"; Flags: onlyifdoesntexist
Source: "release\assets\*"; DestDir: "{app}\assets\"

[Components]
Name: "main"; Description: "Main Files"; Types: full compact custom; Flags: fixed

[Setup]
AppName=Huntstats
AppVersion=0.1.5
AppPublisher=Scopestats
WizardStyle=modern
UsePreviousAppDir=yes
DefaultDirName={autopf}\huntstats
DefaultGroupName=Huntstats
UninstallDisplayIcon={app}\huntstats.exe
Compression=lzma2
SolidCompression=yes
OutputBaseFilename=huntstats_installer
OutputDir="release"

[Tasks]
Name: desktopicon; Description: "Create a &desktop icon"; GroupDescription: "Additional icons:"