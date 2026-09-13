; Inno Setup script for Hettix. Packages the Hettix binary together with the
; bundled portable browser (fetched by scripts/fetch-browser.ps1) into a single
; installer that installs both into the program folder, so the launcher finds
; the browser next to the executable.
;
; Build with: ISCC.exe installer\hettix.iss  (see scripts/build-installer.ps1)

#define MyAppName "Hettix"
#define MyAppExeName "Hettix.exe"
#ifndef MyAppVersion
  #define MyAppVersion "0.0.0"
#endif

[Setup]
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher=Vibe-Coding-Base
DefaultDirName={autopf}\Hettix
DefaultGroupName=Hettix
DisableProgramGroupPage=yes
OutputDir=..\releases
OutputBaseFilename=hettix-setup-{#MyAppVersion}
Compression=lzma2
SolidCompression=yes
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
WizardStyle=modern

[Tasks]
Name: "desktopicon"; Description: "Create a desktop shortcut"; GroupDescription: "Additional icons:"

[Files]
Source: "..\releases\Hettix.exe"; DestDir: "{app}"; Flags: ignoreversion
; The bundled portable browser. Fetch it first with scripts/fetch-browser.ps1.
Source: "..\browser\*"; DestDir: "{app}\browser"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\Hettix"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Uninstall Hettix"; Filename: "{uninstallexe}"
Name: "{autodesktop}\Hettix"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Launch Hettix now"; Flags: nowait postinstall skipifsilent
