# Configuration
$BASE_URL = "https://common.repositories.cloud.sap/artifactory/portal/go/plugins/goli/"
$USERNAME = "i564168"
$PASSWORD = "cmVmdGtuOjAxOjE3NTUwMDU0NjE6bG11dngyWEZINEV5QmVrZUhYZ2k2NHZ6Q3NK"
$VERSION_FILE = "version.txt"
$LATEST_VERSION_FILE = "latest_version.txt"
$OUTPUT_FILE = "Goli-latest-version.zip"
$auth = [System.Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${USERNAME}:${PASSWORD}"))
# Function to get data from Artifactory
function Get-DataFromArtifactory {
    param (
        [string]$fileName
    )
    $url = "$BASE_URL$fileName"
    Invoke-RestMethod -Uri $url -Headers @{Authorization=("Basic {0}" -f $auth)} -Method Get -UseBasicParsing
}

# Read the current version
if (-Not (Test-Path $VERSION_FILE)) {
    $currentVersion = "0.0.0"
} else {
    $currentVersion = Get-Content -Path $VERSION_FILE -Raw
}

# Get the latest version from Artifactory
$version = (Get-DataFromArtifactory $LATEST_VERSION_FILE).Trim()

# Compare versions
if ($version -eq $currentVersion) {
    Write-Host "Already up to date."
    exit 0
}

# Determine OS and Architecture
$OS = (Get-WmiObject Win32_OperatingSystem).Caption
$ARCH = (Get-WmiObject Win32_Processor).AddressWidth
$fileToDownload = ""

if ($OS -like "*Windows*") {
    if ($ARCH -eq 64) {
        $fileToDownload = "Goli-$version-windows-amd64.zip"
    }
} elseif ($OS -like "*macOS*") {
    $fileToDownload = "Goli-$version-macOS-arm64.zip"
} elseif ($OS -like "*Linux*") {
    if ($ARCH -eq 64) {
        $fileToDownload = "Goli-$version-linux-amd64.zip"
    } else {
        $fileToDownload = "Goli-$version-linux-arm64.zip"
    }
}

if ($fileToDownload -eq "") {
    exit 1
}

# Download the latest version
Invoke-RestMethod -Uri "$BASE_URL$fileToDownload" -OutFile $OUTPUT_FILE -Headers @{Authorization=("Basic {0}" -f $auth)}

if ($?) {
    Write-Host "Downloaded and saved as $OUTPUT_FILE"
} else {
    exit 1
}

# Unzip the file
Expand-Archive -Path $OUTPUT_FILE -DestinationPath "./" -Force
Copy-Item -Path "./goliCli/*" -Destination "./" -Recurse -Force
Remove-Item -Recurse -Force "./goliCli"
Remove-Item $OUTPUT_FILE
