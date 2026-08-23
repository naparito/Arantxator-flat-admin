# installer/

Script de [Inno Setup](https://jrsoftware.org/isinfo.php)
([`setup.iss`](setup.iss)) que empaqueta `bin/arantxator.exe` en un
instalador de Windows de doble clic (`Arantxator-Setup.exe`), más el
icono de la aplicación ([`icon.ico`](icon.ico)).

El instalador solo copia el ejecutable y crea los accesos directos — no
descarga nada, porque el frontend, la base de datos y toda la lógica ya
viajan embebidos dentro del propio binario.

Documentación completa del proceso, incluida la verificación realizada:
[`docs/despliegue/instalacion-despliegue.md`](../docs/despliegue/instalacion-despliegue.md).

Resumen rápido:

```powershell
# Todo en un comando (SPA + binario + instalador):
powershell -File ../scripts/build.ps1

# O solo el instalador, si bin/arantxator.exe ya está compilado:
& "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe" setup.iss
```

`icon.ico` se genera con [`../scripts/generate-icon.ps1`](../scripts/generate-icon.ps1)
a partir del isotipo de la marca; solo hace falta volver a ejecutarlo si
cambia el diseño.
