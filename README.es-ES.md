

# ![PersonalMediaVault](./favicon.readme.png) Bóveda de Medios Personales

[![Backend](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/backend.yml/badge.svg)](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/backend.yml)
[![Frontend](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/frontend.yml/badge.svg)](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/frontend.yml)
[![Launcher (CLI)](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/launcher.yml/badge.svg)](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/launcher.yml)
[![Launcher (GUI)](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/launcher-gui.yml/badge.svg)](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/launcher-gui.yml)
[![Backup tool](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/backup-tool.yml/badge.svg)](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/backup-tool.yml)
[![SSE](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/sse.yml/badge.svg)](https://github.com/AgustinSRG/PersonalMediaVault/actions/workflows/sse.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](./LICENSE)

Aplicación web autoalojada para almacenar archivos de medios (vídeos, audio e imágenes) en un almacenamiento cifrado, y visualizarlos utilizando un navegador web.

![Screenshot](https://agustinsrg.github.io/pmv-site/images/screenshot.png)

### Enlaces de acceso rápido

- 🌐 [Sitio web del producto PersonalMediaVault](https://agustinsrg.github.io/pmv-site/)
- 📣 [Anuncios](https://github.com/AgustinSRG/PersonalMediaVault/discussions/categories/announcements)
- 🐛 [Informar de un error](https://github.com/AgustinSRG/PersonalMediaVault/issues/new?template=bug_report.md)
- ☣️ [Informar de una vulnerabilidad](https://github.com/AgustinSRG/PersonalMediaVault/security/advisories/new)
- 💡 [Sugerir una nueva función](https://github.com/AgustinSRG/PersonalMediaVault/issues/new?template=feature_request.md)
- 🙏 [¿Necesitas ayuda? Haz una pregunta](https://github.com/AgustinSRG/PersonalMediaVault/discussions/new?category=q-a)

## Características

- Compatible con vídeos, audio e imágenes.
- Interfaz web: El proyecto proporciona una interfaz web, lo que permite acceder a la bóveda desde múltiples dispositivos y sistemas operativos.
- Almacenamiento cifrado y fácil de respaldar: La bóveda está cifrada usando la contraseña del usuario, y se almacena en archivos dentro de una carpeta de datos. Es posible realizar copias de seguridad de la bóveda simplemente copiando la carpeta, ideal para herramientas como rsync. Este proyecto también proporciona su propia herramienta de respaldo para ello.
- Etiquetas indexadas: Cada recurso multimedia puede etiquetarse, para que el usuario pueda buscarlo con un sistema de búsqueda basado en etiquetas.
- Álbumes: Los archivos multimedia pueden organizarse utilizando álbumes.
- Codificación de medios: Los archivos multimedia se codifican en múltiples formatos para permitir la reproducción desde varios navegadores. Para vídeos e imágenes, también se pueden redimensionar a múltiples resoluciones para diferentes tipos de dispositivos.
- Generación automática de miniaturas: Cuando se suben vídeos o imágenes, se genera una miniatura predeterminada. Puedes cambiar la miniatura predeterminada en cualquier momento después de la subida.
- Compatible con subtítulos: Se pueden añadir subtítulos SubRip a tus vídeos y audios.
- Compatible con múltiples pistas de audio: Los vídeos pueden tener varias pistas de audio, generalmente para audio en varios idiomas.
- Vistas previas de la línea de tiempo del vídeo: Obtén vistas previas al pasar el cursor sobre la línea de tiempo del vídeo.
- Segmentos de línea de tiempo del vídeo: Divide los vídeos largos en secciones para navegar por ellos rápidamente.
- Anotaciones de imágenes: Añade comentarios a las imágenes.
- Adjuntos: Sube cualquier archivo adjunto a tus medios para conservarlos en la bóveda cifrada. Esto se puede utilizar para conservar una copia del archivo multimedia sin procesar, antes de la codificación.
- Búsqueda semántica: Utiliza un modelo de embeddings de código abierto para realizar búsquedas semánticas en tus imágenes.

## Ejecutar en un contenedor

Ejecutar la bóveda en un contenedor aislado es la forma más segura de ejecutarla, da acceso a todas las funciones y permite una configuración detallada. 

Sin embargo, esto puede ser un poco complejo para usuarios no técnicos. Por lo tanto, si solo necesitas una bóveda local, ve a la sección [Paquetes de instalación](#installation-packages) en su lugar.

Necesitarás un sistema de gestión de contenedores, por ejemplo [Docker](https://www.docker.com/) o [Podman](https://podman.io/). En la documentación se utilizan comandos de Docker, por lo que, si estás usando Podman, asegúrate de reemplazar `docker` por `podman` antes de ejecutarlos.

Puedes encontrar la imagen oficial del proyecto subida a [Docker Hub](https://hub.docker.com/r/asanrom/pmv) y [GitHub Packages](https://github.com/AgustinSRG/PersonalMediaVault/pkgs/container/personalmediavault).

Para crear una configuración de bóveda, primero crea un archivo compose ([docker-compose.yml](./packages/docker-compose/docker-compose.yml)):

<details>
<summary>Contenido del archivo (docker-compose.yml):</summary>

```yml
services:
  pmvd:
    hostname: "pmvd"
    image: "asanrom/pmv"
    ports:
      - "${VAULT_PORT}:8000"
    restart: unless-stopped
    volumes:
      - ${VAULT_PATH:-./vault}:/vault
      - ${VAULT_SSL_PATH:-./ssl}:/ssl:ro
      - ${SSE_MODEL_PATH:-./open-clip-model}:/open-clip-model:ro
    environment:
      - USING_PROXY=${USING_PROXY:-NO}
      - VAULT_INITIAL_USER=${VAULT_INITIAL_USER:-admin}
      - VAULT_INITIAL_PASSWORD=${VAULT_INITIAL_PASSWORD:-changeme}
      - SSL_CERT=${SSL_CERT:-}
      - SSL_KEY=${SSL_KEY:-}
      - SEMANTIC_SEARCH_ENABLED=${SEMANTIC_SEARCH_ENABLED:-NO}
      - SSE_MODEL_PATH=/open-clip-model
      - SSE_IMAGE_SIZE_LIMIT_MB=${SSE_IMAGE_SIZE_LIMIT_MB:-20}
    command:
      --daemon
      --clean
      --port 8000
      --skip-lock
      --vault-path /vault
      --cache-size ${VAULT_CACHE_SIZE:-1024}
      ${VAULT_EXTRA_OPTIONS:-}
```
</details><br>

Después, crea en la misma carpeta un archivo de entorno ([.env](./packages/docker-compose/.env.example)) para los parámetros de configuración del archivo compose.

<details>
<summary>Contenido del archivo (.env):</summary>

```sh
####################################
# PersonalMediaVault configuration #
####################################

# Listening port
VAULT_PORT=8000

# Path where the vault will be stored
VAULT_PATH=./vault

# SSL (recommended)
#
# Running the vault with HTTP is the most secure option
# Obtain a key and a certificate for your domain
#
# Set VAULT_SSL_PATH to the path where the key and the certificate are stored
# This path will be mapped to /ssl in the container
# Set SSL_KEY and SSL_CERT to the key and certificate chain files respectively
# The files must be in the /ssl path (eg: /ssl/certificate.pem) and in PEM format

VAULT_SSL_PATH=./ssl

#SSL_KEY=/ssl/key.pem
#SSL_CERT=/ssl/certificate.pem

# Reverse proxy
#
# Sometimes is better to use a reverse proxy (line NGINX)
# as the frontend and forward the requests to the daemon
#
# If you are using it, set USING_PROXY to YES in order for
# the daemon to change the way it resolves the IP addresses of clients

USING_PROXY=NO

# Initial vault user
#
# If the vault has no users, an initial user will be created
#
# Set VAULT_INITIAL_USER and VAULT_INITIAL_PASSWORD 
# for the username and password respectively
#
# Make sure to change them the first time you log into the vault.

VAULT_INITIAL_USER=admin
VAULT_INITIAL_PASSWORD=changeme

# Cache size
# You can modify it to accelerate the read speed
# but will also result in higher memory usage
# The recommended value is 1024

VAULT_CACHE_SIZE=1024

# Extra options
#
# You can set the following extra options,
# separating them with spaces:
#
#   --log-requests - Enables request logging
#   --debug - Enables debug logging (useful for troubleshooting)
#   --check-trash - Checks the vault (at startup) in order to find trash files. This option requires the vault credentials passed in the environment variables 'VAULT_USER' and 'VAULT_PASSWORD', in order to decrypt the vault files.
#   --remove-trash - Removes the trash files. Combine this option with '--check-trash'.
#   --recover - Recovers non-indexed media assets.

VAULT_EXTRA_OPTIONS=--log-requests

# Enable semantic search?
# This option can be YES or NO
# If set to YES, make sure to also set the model
SEMANTIC_SEARCH_ENABLED=NO

# OpenCLIP model path
# First, download a model (you can find models in Hugging face)
# Hugging face search link: https://huggingface.co/models?pipeline_tag=zero-shot-image-classification&library=onnx&other=clip
# Change this variable to point to the model folder
SSE_MODEL_PATH=./open-clip-model

# Max size for images before they cannot be longer encoded
# The size is set in MegaBytes (no decimals allowed)
# This limits the memory usage of the daemon
# If you plan to work with very large images, make sure to set this to a high value
SSE_IMAGE_SIZE_LIMIT_MB=20
```
</details><br>

Lee el archivo de entorno para configurarlo. Luego, puedes iniciar la bóveda:

```sh
docker compose up -d
```

En caso de que quieras actualizar la imagen, ejecuta:

```sh
docker compose pull
docker compose up -d
```

En caso de que quieras detenerla, ejecuta:

```sh
docker compose down
```

## Paquetes de instalación

Ofrecemos varios paquetes de instalación dependiendo del sistema operativo que estés utilizando. Puedes encontrar esos paquetes en la sección [Lanzamientos](https://github.com/AgustinSRG/PersonalMediaVault/releases) del repositorio.

Este método es ideal para usuarios no técnicos, ya que se instalará un lanzador gráfico para ejecutar la bóveda sin necesidad de comandos.

Lee las secciones siguientes para ver las instrucciones de instalación.

### Windows

Para Windows, descarga el instalador **MSI** (`PersonalMediaVault-{VERSION}-x64.msi`) desde [Lanzamientos](https://github.com/AgustinSRG/PersonalMediaVault/releases).

Ejecuta el instalador y todo se configurará para que la aplicación funcione.

Una vez instalado, crea una carpeta vacía para almacenar tu bóveda de medios y haz clic derecho en ella. Deberías ver una nueva opción "Open with PersonalMediaVault". Haz clic en ella para ejecutar el lanzador de la bóveda.

### Debian, Ubuntu u otra distribución de Linux basada en Debian

Para distribuciones de Linux basadas en Debian, descarga el paquete **DEB** (`personalmediavault_{VERSION}_amd64.deb`) desde [Lanzamientos](https://github.com/AgustinSRG/PersonalMediaVault/releases).

Instala el paquete, con `apt` o con tu gestor de paquetes gráfico.

También puedes instalarlo directamente desde el repositorio PPA (construido automáticamente con GitHub actions):

```sh
# Download the public key
curl -s --compressed "https://agustinsrg.github.io/PersonalMediaVault/KEY.gpg" | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/pmv.gpg >/dev/null

# Add the APT list file
sudo curl -s --compressed -o /etc/apt/sources.list.d/pmv.list "https://agustinsrg.github.io/PersonalMediaVault/pmv.list"

# Update APT lists
sudo apt update

# Install the package
sudo apt install personalmediavault
```

Una vez instalado, puedes ejecutar el lanzador con `pmv /path/to/vault` o el lanzador gráfico buscando "PersonalMediaVault" en tu menú de aplicaciones.

### Fedora Linux

Para distribuciones de Linux basadas en Fedora, descarga el paquete **RPM** (`personalmediavault-{VERSION}.x86_64.rpm`) desde [Lanzamientos](https://github.com/AgustinSRG/PersonalMediaVault/releases).

Instala el paquete, con `dnf` o con tu gestor de paquetes gráfico.

Una vez instalado, puedes ejecutar el lanzador con `pmv /path/to/vault` o el lanzador gráfico buscando "PersonalMediaVault" en tu menú de aplicaciones.

### Arch Linux

Para distribuciones de Linux basadas en Arch, descarga el paquete **PKG.TAR.ZST** (`personalmediavault-{VERSION}-x86_64.pkg.tar.zst`) desde [Lanzamientos](https://github.com/AgustinSRG/PersonalMediaVault/releases).

Instala el paquete, con `pacman`:

```sh
sudo pacman -U personalmediavault-{VERSION}-x86_64.pkg.tar.zst
```

Una vez instalado, puedes ejecutar el lanzador con `pmv /path/to/vault` o el lanzador gráfico buscando "PersonalMediaVault" en tu menú de aplicaciones.

### Otras distribuciones de Linux

Para cualquier otra distribución de Linux, puedes descargar el paquete **TAR.GZ** (`personalmediavault_{VERSION}_amd64.tar.gz`) desde [Lanzamientos](https://github.com/AgustinSRG/PersonalMediaVault/releases).

Descomprímelo y ejecuta el script de instalación (`install.sh`) con privilegios de administración:

```sh
sudo ./install.sh
```

Una vez instalado, puedes ejecutar el lanzador con `pmv /path/to/vault` o el lanzador gráfico buscando "PersonalMediaVault" en tu menú de aplicaciones.

### Menú de servicios de KDE

Si estás utilizando una distribución de Linux y KDE como entorno de escritorio, puedes obtener las mismas opciones del menú contextual para abrir bóvedas que se instalan en el paquete de Windows.

Crea el archivo `~/.local/share/kio/servicemenus/pmv-open.desktop` con el siguiente contenido:

<details>
<summary>Contenido del archivo (pmv-open.desktop):</summary>

```conf
[Desktop Entry]
Name=Open with PersonalMediaVault
Type=Service
ServiceTypes=KonqPopupMenu/Plugin
MimeType=inode/directory
Actions=openWithPmv

[Desktop Action openWithPmv]
Name=Open with PersonalMediaVault
Name[es]=Abrir con PersonalMediaVault
Terminal=false
Icon=pmv
StartupWMClass=pmv
Exec=pmv-gui .
```
</details><br>

Si prefieres utilizar el lanzador de consola en su lugar, usa la siguiente entrada de escritorio:

<details>
<summary>Contenido del archivo (pmv-open-console.desktop):</summary>

```conf
[Desktop Entry]
Name=Open with PersonalMediaVault (Console)
Type=Service
ServiceTypes=KonqPopupMenu/Plugin
MimeType=inode/directory
Actions=openWithPmvConsole

[Desktop Action openWithPmvConsole]
Name=Open with PersonalMediaVault (Console)
Name[es]=Abrir con PersonalMediaVault (Consola)
Terminal=false
Icon=pmv
StartupWMClass=pmv
Exec=konsole --separate -e 'pmv .'
```
</details>

## Componentes del proyecto

- [Backend](./backend): Gestiona la bóveda, codifica y almacena los archivos multimedia y proporciona una API HTTP para que los clientes accedan a ella.
- [Frontend](./frontend): Utiliza la API HTTP para acceder a la bóveda, proporcionando una interfaz web al usuario.
- [Herramienta de respaldo](./backup-tool): Herramienta para realizar copias de seguridad de bóvedas multimedia (copia los archivos nuevos y reemplaza los antiguos, utilizando la fecha de última modificación).
- [Lanzador](./launcher): Programa CLI para iniciar la aplicación web. Este componente se ha creado para facilitar su uso en local, cuando no tienes el backend configurado como un servicio del sistema.
- [Motor de búsqueda semántica](./semantic-search-engine): Un servidor interno para ejecutar modelos de embeddings con el fin de realizar la funcionalidad de búsqueda semántica.
- [Paquetes de instalación](./packages): Colección de paquetes de instalación para múltiples sistemas operativos.

## Motivación del proyecto

Este es un proyecto personal desarrollado con el objetivo de crear una herramienta de gestión multimedia que tenga una interfaz web similar a la de YouTube, manteniendo al mismo tiempo los archivos multimedia cifrados en disco.

**¿Por qué cifrado?:** Al almacenar tus imágenes, vídeos o grabaciones de audio personales, si los guardas sin cifrar en tu computadora, cualquier programa malicioso capaz de infectar tu dispositivo podrá obtenerlos fácilmente del sistema de archivos. Al cifrarlos, se dificulta que esos archivos multimedia sean robados. Además, al estar cifrados, puedes realizar copias de seguridad de ellos fácilmente en servicios de almacenamiento en la nube como Google Drive, sin darle a Google la posibilidad de espiar tus archivos multimedia personales.

**¿Por qué una aplicación web?:** Al hacerla una aplicación web, su uso se vuelve muy flexible. Puedes ejecutar el backend en un ordenador de tu LAN o en un servidor remoto y podrás usar la app desde todos los dispositivos conectados a Internet sin necesidad de instalaciones adicionales. Además, las interfaces web son más fáciles de desarrollar para funcionar en múltiples dispositivos y ofrecen características estándar ya implementadas que son necesarias para la aplicación, como el reproductor de vídeo.

## Licencia

Este proyecto está bajo la [Licencia MIT](./LICENSE).

## Contribuir

- [Solicitudes de extracción (Pull requests)](https://github.com/AgustinSRG/PersonalMediaVault/pulls)
- [Normas de contribución](./CONTRIBUTING.md)
