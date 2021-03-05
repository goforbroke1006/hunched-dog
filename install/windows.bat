
@echo "hunched-dog installation..."


if not exist "%ProgramFiles%\hunched-dog\" (
    echo 'Create installation directory' "%ProgramFiles%\hunched-dog\"
    mkdir "%ProgramFiles%\hunched-dog\"
)

CALL curl -L https://github.com/goforbroke1006/hunched-dog/releases/download/v0.1.4/hunched-dog__windows_amd64.exe --output "%ProgramFiles%\hunched-dog\hunched-dog.exe"


:: sc.exe create "hunched-dog" binPath= "%ProgramFiles%\hunched-dog\hunched-dog.exe"
:: echo 'Create windows service'


if not exist "%UserProfile%\hunched-dog-cloud" (
    echo 'Create default sync directory'
    mkdir "%UserProfile%\hunched-dog-cloud\"
)


if not exist "%ProgramFiles%\hunched-dog\config.yml" (
    (
        @echo target: %UserProfile%/hunched-dog-cloud
        @echo multicast: 224.0.0.0:45046
    ) > "%ProgramFiles%\hunched-dog\config.yml"
)

if not exist "%UserProfile%\.hunched-dog\" (
    echo 'Create configs directory'
    mkdir "%UserProfile%\.hunched-dog\"
)

if not exist "%UserProfile%\.hunched-dog\config.yml" (
    (
        @echo target: %UserProfile%/hunched-dog-cloud
        @echo multicast: 224.0.0.0:45046
    ) > "%UserProfile%\.hunched-dog\config.yml"
)

:: shortcut /a:c /f:"%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\hunched-dog.lnk" /t:"%ProgramFiles%\hunched-dog\hunched-dog.exe"
:: powershell -Command "New-Item -ItemType SymbolicLink -Path \"%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\" -Name \"hunched-dog.lnk\" -Value \"%ProgramFiles%\hunched-dog\hunched-dog.exe\" -Force"

SET LONG_COMMAND= ^
$WshShell = New-Object -comObject WScript.Shell; ^
$Shortcut = $WshShell.CreateShortcut('"%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\hunched-dog.lnk"'); ^
$Shortcut.TargetPath = '"%ProgramFiles%\hunched-dog\hunched-dog.exe"'; ^
$Shortcut.Arguments = '"argumentA ArgumentB"'; ^
$Shortcut.WorkingDirectory = '"%UserProfile%\.hunched-dog"'; ^
$Shortcut.Save()

START Powershell -noexit -command "%LONG_COMMAND%"

echo 'Add to startup applications list' "%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\hunched-dog.lnk"

start "%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\hunched-dog.lnk"
:: Start-Process java -ArgumentList '-jar', 'MyProgram.jar' -RedirectStandardOutput '.\console.out' -RedirectStandardError '.\console.err'

PAUSE