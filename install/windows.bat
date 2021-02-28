@echo off

(New-Object Net.WebClient).DownloadFile('https://github.com/goforbroke1006/hunched-dog/releases/download/0.1.0/hunched-dog__windows_amd64', 'hunched-dog.exe')
if not exist "%ProgramFiles%\hunched-dog\" (
    echo 'Create installation directory'
    mkdir %ProgramFiles%\hunched-dog\
)
xcopy /s .\hunched-dog.exe "%ProgramFiles%\hunched-dog\hunched-dog.exe"
sc.exe create "hunched-dog" binPath= "%ProgramFiles%\hunched-dog\hunched-dog.exe"
echo 'Create windows service'


if not exist "%UserProfile%\hunched-dog-cloud" (
    echo 'Create default sync directory'
    mkdir %UserProfile%\hunched-dog-cloud\
)

if not exist "%UserProfile%\.hunched-dog\" (
    echo 'Create configs directory'
    mkdir %UserProfile%\.hunched-dog\
)

if not exist "%UserProfile%\.hunched-dog\config.yml" (
    (
        @echo "target: %UserProfile%/hunched-dog-cloud"
        @echo "hosts:"
        @echo "  - 192.168.0.1"
        @echo "  - 192.168.0.2"
        @echo "  - 192.168.0.3"
        @echo "  - 192.168.0.4"
        @echo "  - 192.168.0.5"
        @echo "  - 192.168.0.6"
        @echo "  - 192.168.0.7"
        @echo "  - 192.168.0.8"
        @echo "  - 192.168.0.9"
        @echo "  - 192.168.0.10"
        @echo "  - 192.168.0.88"
    ) > "%UserProfile%\.hunched-dog\config.yml"
)