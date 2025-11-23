const fs = require('fs');
const path = require('path');

const versionFilePath = path.join(__dirname, '../VERSION');
const currentVersion = fs.readFileSync(versionFilePath, 'utf8').trim();

const versionParts = currentVersion.split('.').map(Number);
versionParts[2] += 1;
const newVersion = versionParts.join('.');

fs.writeFileSync(versionFilePath, newVersion);

console.log(newVersion);
