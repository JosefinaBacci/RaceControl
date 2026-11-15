"""Builds a PowerPoint with each page of the PDF as a full-slide image (same look as the PDF, not editable).
Usage: pdf2pptx.py <in.pdf> <out.pptx>"""
import subprocess, sys, tempfile
from pathlib import Path
from pptx import Presentation
from pptx.util import Inches

pdf, out = sys.argv[1], sys.argv[2]

with tempfile.TemporaryDirectory() as tmp:
    # 1920 px wide: sharp on a Full HD projector
    subprocess.run(['pdftoppm', '-png', '-scale-to-x', '1920', '-scale-to-y', '-1', pdf, f'{tmp}/slide'], check=True)
    deck = Presentation()
    deck.slide_width, deck.slide_height = Inches(13.333), Inches(7.5)  # 16:9, same as the PDF
    blank = deck.slide_layouts[6]
    for image in sorted(Path(tmp).glob('slide-*.png')):
        deck.slides.add_slide(blank).shapes.add_picture(str(image), 0, 0, deck.slide_width, deck.slide_height)
    deck.save(out)
    print(f'{out}: {len(deck.slides)} slides')
