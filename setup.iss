; ============================================
; Inno Setup Script for LockScreen Sync
; ============================================

#define MyAppName "LockScreen Sync"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "LockScreen Sync"
#define MyAppExeName "LockScreenSync.exe"

[Setup]
; Основні налаштування
AppId={{A1B2C3D4-E5F6-7890-ABCD-EF1234567890}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
OutputDir=installer
OutputBaseFilename=LockScreenSync_Setup
SetupIconFile=icon.ico
UninstallDisplayIcon={app}\{#MyAppExeName}
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
ArchitecturesInstallIn64BitMode=x64compatible

; Мова
LanguageDetectionMethod=uilanguage
ShowLanguageDialog=auto

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "ukrainian"; MessagesFile: "compiler:Languages\Ukrainian.isl"

[Tasks]
Name: "autostart"; Description: "Запускати при старті Windows"; GroupDescription: "Додаткові опції:"; Flags: checkedonce
Name: "desktopicon"; Description: "Створити ярлик на робочому столі"; GroupDescription: "Додаткові опції:"; Flags: unchecked

[Files]
Source: "LockScreenSync.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
; Меню Пуск
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Видалити {#MyAppName}"; Filename: "{uninstallexe}"
; Робочий стіл (опціонально)
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Registry]
; Реєстр не використовується для автозавантаження - використовуємо Task Scheduler

[Run]
; Запустити після встановлення
Filename: "{app}\{#MyAppExeName}"; Description: "Запустити {#MyAppName}"; Flags: nowait postinstall skipifsilent runascurrentuser

[UninstallRun]
; Закрити програму перед видаленням
Filename: "taskkill"; Parameters: "/F /IM {#MyAppExeName}"; Flags: runhidden

[UninstallDelete]
; Очистити залишкові файли
Type: files; Name: "{app}\*"
Type: dirifempty; Name: "{app}"

[Code]
// Перевірка чи програма вже запущена
function IsAppRunning(): Boolean;
var
  ResultCode: Integer;
begin
  Exec('tasklist', '/FI "IMAGENAME eq LockScreenSync.exe" /NH', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Result := (ResultCode = 0);
end;

// Створити завдання Task Scheduler для автозавантаження з правами адміністратора
procedure CreateScheduledTask(ExePath: String);
var
  ResultCode: Integer;
  TaskXML: String;
  TempFile: String;
begin
  TempFile := ExpandConstant('{tmp}\LockScreenSyncTask.xml');

  TaskXML := '<?xml version="1.0" encoding="UTF-16"?>' + #13#10 +
    '<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">' + #13#10 +
    '  <RegistrationInfo>' + #13#10 +
    '    <Description>LockScreen Sync - синхронізація шпалер з екраном блокування</Description>' + #13#10 +
    '  </RegistrationInfo>' + #13#10 +
    '  <Triggers>' + #13#10 +
    '    <LogonTrigger>' + #13#10 +
    '      <Enabled>true</Enabled>' + #13#10 +
    '    </LogonTrigger>' + #13#10 +
    '  </Triggers>' + #13#10 +
    '  <Principals>' + #13#10 +
    '    <Principal id="Author">' + #13#10 +
    '      <LogonType>InteractiveToken</LogonType>' + #13#10 +
    '      <RunLevel>HighestAvailable</RunLevel>' + #13#10 +
    '    </Principal>' + #13#10 +
    '  </Principals>' + #13#10 +
    '  <Settings>' + #13#10 +
    '    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>' + #13#10 +
    '    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>' + #13#10 +
    '    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>' + #13#10 +
    '    <AllowHardTerminate>true</AllowHardTerminate>' + #13#10 +
    '    <StartWhenAvailable>true</StartWhenAvailable>' + #13#10 +
    '    <RunOnlyIfNetworkAvailable>false</RunOnlyIfNetworkAvailable>' + #13#10 +
    '    <AllowStartOnDemand>true</AllowStartOnDemand>' + #13#10 +
    '    <Enabled>true</Enabled>' + #13#10 +
    '    <Hidden>false</Hidden>' + #13#10 +
    '    <RunOnlyIfIdle>false</RunOnlyIfIdle>' + #13#10 +
    '    <WakeToRun>false</WakeToRun>' + #13#10 +
    '    <ExecutionTimeLimit>PT0S</ExecutionTimeLimit>' + #13#10 +
    '    <Priority>7</Priority>' + #13#10 +
    '  </Settings>' + #13#10 +
    '  <Actions Context="Author">' + #13#10 +
    '    <Exec>' + #13#10 +
    '      <Command>' + ExePath + '</Command>' + #13#10 +
    '    </Exec>' + #13#10 +
    '  </Actions>' + #13#10 +
    '</Task>';

  SaveStringToFile(TempFile, TaskXML, False);

  // Видаляємо старе завдання якщо існує
  Exec('schtasks', '/Delete /TN "LockScreenSync" /F', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);

  // Створюємо нове завдання
  Exec('schtasks', '/Create /TN "LockScreenSync" /XML "' + TempFile + '"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);

  DeleteFile(TempFile);
end;

// Видалити завдання Task Scheduler
procedure DeleteScheduledTask();
var
  ResultCode: Integer;
begin
  Exec('schtasks', '/Delete /TN "LockScreenSync" /F', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
end;

// Закрити програму перед встановленням/оновленням
procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode: Integer;
begin
  if CurStep = ssInstall then
  begin
    Exec('taskkill', '/F /IM LockScreenSync.exe', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    Sleep(500);
  end;

  // Після встановлення створюємо завдання планувальника
  if CurStep = ssPostInstall then
  begin
    if IsTaskSelected('autostart') then
    begin
      CreateScheduledTask(ExpandConstant('{app}\{#MyAppExeName}'));
    end;
  end;
end;

// Закрити програму перед видаленням
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  ResultCode: Integer;
begin
  if CurUninstallStep = usUninstall then
  begin
    Exec('taskkill', '/F /IM LockScreenSync.exe', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    DeleteScheduledTask();
    Sleep(500);
  end;
end;
