# Defensa final

Diapositivas de la defensa del proyecto Fly Away, desde el punto de vista del grupo
planificador. Están escritas en Markdown (`presentacion.md`) y se arman con Pandoc + Beamer,
con el tema Nord de `nord.tex`.

```bash
make -C docs/defensa        # genera presentacion.pdf
make -C docs/defensa pptx   # genera también presentacion.pptx
```

El PowerPoint tiene cada diapositiva como imagen: se ve igual que el PDF pero no se puede
editar ahí. Los cambios se hacen en `presentacion.md` y se vuelve a correr `make`.

Requiere `pandoc`, TeX Live con LuaLaTeX (paquetes `beamer`, `metropolis`, `tcolorbox`,
`fontawesome5` y las fuentes Fira) y, para el PowerPoint, `pdftoppm` (poppler) y uv.

No dejes que un formateador de Markdown (Prettier, formato al guardar) toque
`presentacion.md`: no entiende los bloques `:::` de Pandoc y rompe las diapositivas.
