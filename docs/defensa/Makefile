presentacion.pdf: presentacion.md nord.tex beamer.lua
	pandoc presentacion.md -t beamer --pdf-engine=lualatex -V fontsize=10pt -H nord.tex -L beamer.lua -o $@

# PowerPoint with each slide as an image (LibreOffice's PDF import scrambles the text)
pptx: presentacion.pptx
presentacion.pptx: presentacion.pdf pdf2pptx.py
	uv run --no-project --with python-pptx python pdf2pptx.py $< $@

.PHONY: pptx
