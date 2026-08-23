; Script de Inno Setup para Arantxator Flat Admin.
;
; Empaqueta el ejecutable ya compilado (bin/arantxator.exe, que ya lleva
; embebidos la SPA, el motor SQLite y todos los datos que necesita) en un
; instalador de doble clic autocontenido: no descarga nada durante la
; instalacion porque no hay nada que descargar, todo va dentro del propio
; instalador.
;
; Requiere Inno Setup 6 (https://jrsoftware.org/isinfo.php) y que
; bin/arantxator.exe ya este compilado (ver scripts/build.ps1, que llama a
; este script como ultimo paso).
;
; Compilacion manual:
;   "C:\...\Inno Setup 6\ISCC.exe" installer\setup.iss

#define MyAppName "Arantxator Flat Admin"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "Arantxator"
#define MyAppExeName "arantxator.exe"

[Setup]
AppId={{6F2B6A7E-6E6E-4B1E-9C5B-5B6E6E1E9A11}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
; No hace falta privilegios de administrador: se instala en el perfil del
; usuario actual, sin tocar carpetas del sistema.
PrivilegesRequired=lowest
OutputDir=..\dist
OutputBaseFilename=Arantxator-Setup
SetupIconFile=icon.ico
UninstallDisplayIcon={app}\{#MyAppExeName}
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
; El propio instalador es el unico artefacto que necesita el usuario: nada
; de conexion a internet durante la instalacion.
DisableWelcomePage=no
ArchitecturesInstallIn64BitMode=x64compatible

[Languages]
Name: "spanish"; MessagesFile: "compiler:Languages\Spanish.isl"

[Tasks]
Name: "desktopicon"; Description: "Crear un acceso directo en el &escritorio"; GroupDescription: "Iconos adicionales:"; Flags: unchecked

[Files]
Source: "..\bin\arantxator.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Desinstalar {#MyAppName}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
; Ofrece abrir la aplicacion nada mas terminar el asistente, igual que
; pediria un usuario sin conocimientos tecnicos.
Filename: "{app}\{#MyAppExeName}"; Description: "Abrir {#MyAppName}"; Flags: nowait postinstall skipifsilent
