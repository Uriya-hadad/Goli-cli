const path = require('path');
const axios = require('axios');
const AdmZip = require('adm-zip');

const fs = require('fs');

const username = "i564168";
const password = "cmVmdGtuOjAxOjE3NTUwMDU0NjE6bG11dngyWEZINEV5QmVrZUhYZ2k2NHZ6Q3NK";

const auth = Buffer.from(`${username}:${password}`).toString('base64');
const ARTIFACTORY_URL = "https://common.repositories.cloud.sap/artifactory/portal/go/plugins/goli";

function extractZipToFolder(zipPath) {

    try {
        const zip = new AdmZip(zipPath);

        // Determine the directory of the zip file
        const extractPath = path.dirname(zipPath);

        // Extract the contents to the same folder
        zip.extractAllTo(extractPath, true);

        console.log(`Zip file extracted successfully to: ${extractPath}`);
    } catch (err) {
        console.error('Error extracting zip file:', err);
    }
}

async function getFileFromArtifactoryStream(fileName) {
    const url = `${ARTIFACTORY_URL}/${fileName}`;

    try {
        const response = await axios.get(url, {
            headers: {
                'Authorization': `Basic ${auth}`,
            },
            responseType: 'stream',
        });
        return response.data
    } catch (error) {
        console.error('Error fetching the latest version:', error.message);
        throw new Error('Unable to fetch the latest version');
    }
}
async function getFileFromArtifactory(fileName) {
    const url = `${ARTIFACTORY_URL}/${fileName}`;

    try {
        const response = await axios.get(url, {
            headers: {
                'Authorization': `Basic ${auth}`,
            },
        });
        return response.data
    } catch (error) {
        console.error('Error fetching the latest version:', error.message);
        throw new Error('Unable to fetch the latest version');
    }
}

async function getLatestVersion() {
    return (await getFileFromArtifactory("latest_version.txt"));
}

async function saveMacArtifact(outputPath) {
    version = (await getLatestVersion()).trim();
    return await downloadArtifact(`Goli-${version}-macOS-arm64.zip`, outputPath);
}

async function saveWinArtifact(outputPath) {
    version = (await getLatestVersion()).trim();
    return await downloadArtifact(`Goli-${version}-windows-amd64.zip`, outputPath);
}

async function saveAutoUpdateScriptMac(outputPath) {
    return await downloadArtifact(`autoUpdate.sh`, outputPath);
}

async function saveAutoUpdateScriptWin(outputPath) {
    return await downloadArtifact(`autoUpdate.ps1`, outputPath);
}

async function downloadArtifact(fileName, outputPath) {
    const res = await getFileFromArtifactoryStream(fileName);
    const writer = fs.createWriteStream(outputPath);
    return new Promise((resolve, reject) => {
        res.pipe(writer);
        writer.on('finish', () => {
            console.log(`File downloaded and saved to ${outputPath}`);
            resolve(); // Resolves when file is fully written
        });
        writer.on('error', (err) => {
            console.error('Error while downloading file:', err);
            reject(err);
        });
    })
}

async function createMacPackage(res) {
    try {
        const tempDir = path.join(__dirname, 'temp'+Math.random());
        const outputZipPath = path.join(tempDir, 'final-package.zip');

        // Create a temporary directory for processing
        if (!fs.existsSync(tempDir)) {
            fs.mkdirSync(tempDir);
        }

        // Step 1: Download the ZIP file and bash script as streams
        const zipFilePath = path.join(tempDir, 'downloaded.zip');
        const scriptPath = path.join(tempDir, 'autoUpdate.sh');

        console.log('Downloading files...');
        await saveMacArtifact(zipFilePath);
        await saveAutoUpdateScriptMac(scriptPath);

        // Step 2: Unzip the downloaded ZIP file
        console.log('Extracting files...');
        extractZipToFolder(zipFilePath);

        // Step 3: Add the bash script to the extracted contents
        const finalZip = new AdmZip();
        fs.renameSync(path.join(tempDir, 'autoUpdate.sh'), path.join(tempDir, 'goliCli', 'autoUpdate.sh'));
        
        const goliPath = path.join(tempDir, 'goliCli');
        fs.chmodSync(path.join(goliPath, 'autoUpdate.sh'), 0o755);
        fs.chmodSync(path.join(goliPath, 'goli'), 0o755);
        const extractedFiles = fs.readdirSync(goliPath);
        extractedFiles.forEach((file) => {
            const filePath = path.join(goliPath, file);
            if (file !== 'resources') {
                finalZip.addLocalFile(filePath);
            }
        });
        finalZip.addLocalFolder(path.join(goliPath, 'resources'), 'resources');
        
        // Step 4: Create the new ZIP file
        finalZip.writeZip(outputZipPath);

        // Step 5: Stream the new ZIP file to the client
        res.setHeader('Content-Disposition', 'attachment; filename="goli-macOS-arm64.zip"');
        res.setHeader('Content-Type', 'application/zip');
        const readStream = fs.createReadStream(outputZipPath);
        const a = readStream.pipe(res)
        a.on('finish', () => {
            // Cleanup temporary files
            fs.rmSync(tempDir, {recursive: true, force: true});
        });
        a.on('error', (err) => {
            console.error('Error streaming file:', err);
            res.status(500).send('Error streaming file: ' + err.message);
            fs.rmSync(tempDir, {recursive: true, force: true});
        })
    } catch (error) {
        console.error('Error processing files:', error.message);
        res.status(500).send('Error processing files: ' + error.message);
    }
}

async function createWinPackage(res) {
    try {
        const tempDir = path.join(__dirname, 'temp'+Math.random());
        const outputZipPath = path.join(tempDir, 'final-package.zip');

        // Create a temporary directory for processing
        if (!fs.existsSync(tempDir)) {
            fs.mkdirSync(tempDir);
        }

        // Step 1: Download the ZIP file and bash script as streams
        const zipFilePath = path.join(tempDir, 'downloaded.zip');
        const scriptPath = path.join(tempDir, 'autoUpdate.ps1');

        console.log('Downloading files...');
        await saveWinArtifact(zipFilePath);
        await saveAutoUpdateScriptWin(scriptPath);

        // Step 2: Unzip the downloaded ZIP file
        console.log('Extracting files...');
        extractZipToFolder(zipFilePath);

        // Step 3: Add the bash script to the extracted contents
        const finalZip = new AdmZip();
        fs.renameSync(path.join(tempDir, 'autoUpdate.ps1'), path.join(tempDir, 'goliCli', 'autoUpdate.ps1'));

        const goliPath = path.join(tempDir, 'goliCli');
        fs.chmodSync(path.join(goliPath, 'autoUpdate.ps1'), 0o755);
        fs.chmodSync(path.join(goliPath, 'goli.exe'), 0o755);
        const extractedFiles = fs.readdirSync(goliPath);
        extractedFiles.forEach((file) => {
            const filePath = path.join(goliPath, file);
            if (file !== 'resources') {
                finalZip.addLocalFile(filePath);
            }
        });
        finalZip.addLocalFolder(path.join(goliPath, 'resources'), 'resources');

        // Step 4: Create the new ZIP file
        finalZip.writeZip(outputZipPath);

        // Step 5: Stream the new ZIP file to the client
        res.setHeader('Content-Disposition', 'attachment; filename="goli-windows-amd64.zip"');
        res.setHeader('Content-Type', 'application/zip');
        const readStream = fs.createReadStream(outputZipPath);
        const a = readStream.pipe(res)
        a.on('finish', () => {
            // Cleanup temporary files
            console.log("Finished streaming file");
            console.log('Cleaning up temporary files...');
            fs.rmSync(tempDir, {recursive: true, force: true});
        });
        a.on('error', (err) => {
            console.error('Error streaming file:', err);
            res.status(500).send('Error streaming file: ' + err.message);
            fs.rmSync(tempDir, {recursive: true, force: true});
        })
    } catch (error) {
        console.error('Error processing files:', error.message);
        res.status(500).send('Error processing files: ' + error.message);
    }
}

module.exports = {
    createMacPackage,
    createWinPackage
}
