# installer/

Contendrá el script de [Inno Setup](https://jrsoftware.org/isinfo.php) que
empaqueta `bin/arantxator.exe` en un instalador de Windows de doble clic
(`Arantxator-Setup.exe`).

El instalador solo copia el ejecutable y crea un acceso directo — no
descarga nada, porque el frontend, la base de datos y toda la lógica ya
viajan embebidos dentro del propio binario (ver
[docs/design/diseno-tecnico-funcional.md](../docs/design/diseno-tecnico-funcional.md),
sección "Instalación transparente").

Pendiente de implementar.
