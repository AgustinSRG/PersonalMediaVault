
# Remove old folder
Remove-Item -ErrorAction Ignore -Path "PersonalMediaVault" -Recurse

# Create folder
New-Item -ErrorAction Ignore -Path "PersonalMediaVault" -itemType Directory

# Copy frontend
Copy-Item -Path "..\..\frontend\dist" -Destination "PersonalMediaVault\www" -Recurse -Force

# Create bin folder
New-Item -ErrorAction Ignore -Path "PersonalMediaVault/bin" -itemType Directory

# Copy ffmpeg
Copy-Item -Path "C:\ffmpeg\bin\ffprobe.exe" -Destination "PersonalMediaVault\bin\ffprobe.exe" -Force
Copy-Item -Path "C:\ffmpeg\bin\ffmpeg.exe" -Destination "PersonalMediaVault\bin\ffmpeg.exe" -Force

# Copy backend
Copy-Item -Path "..\..\backend\pmvd.exe" -Destination "PersonalMediaVault\bin\pmvd.exe" -Force

# Copy backup tool
Copy-Item -Path "..\..\backup-tool\pmv-backup.exe" -Destination "PersonalMediaVault\bin\pmv-backup.exe" -Force

# Copy launcher
Copy-Item -Path "..\..\launcher\pmv.exe" -Destination "PersonalMediaVault\pmv.exe" -Force

# Copy launcher (GUI)
Copy-Item -Path "..\..\launcher-gui\target\release\pmv-gui.exe" -Destination "PersonalMediaVault\pmv-gui.exe" -Force
