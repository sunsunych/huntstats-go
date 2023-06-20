#define AppName "Huntstats"
#define AppVersion "0.1.4"
#define AppExeName "huntstats.exe"
#define AppPublisher "ScopesStats"

[Files]
Source: "release\{#AppExeName}"; DestDir: "{app}";
Source: "release\config.toml"; DestDir: "{app}"; Flags: onlyifdoesntexist
Source: "release\assets\*"; DestDir: "{app}\assets\"

[Setup]
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
WizardStyle=modern
UsePreviousAppDir=yes
DefaultDirName={autopf}\huntstats
DefaultGroupName=Huntstats
UninstallDisplayIcon={app}\{#AppExeName}
Compression=lzma2
SolidCompression=yes
OutputBaseFilename=huntstats_installer
OutputDir="release"
SignTool=Huntstats $f

[Icons]
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#AppExeName}"