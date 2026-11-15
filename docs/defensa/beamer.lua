-- Turns fenced divs into the boxes defined in nord.tex and renders tables as plain booktabs tabulars
-- (pandoc's longtable doesn't work inside beamer columns).

local function latex(blocks)
  return (pandoc.write(pandoc.Pandoc(blocks), 'latex'):gsub('%s+$', ''))
end

local boxes = {decision = true, joke = true, result = true, caution = true, experiment = true}
local sizes = {small = true, footnotesize = true, scriptsize = true}

function Div(el)
  for _, class in ipairs(el.classes) do
    if boxes[class] then
      local title = latex(pandoc.read(el.attributes.title or '', 'markdown').blocks)
      local blocks = {pandoc.RawBlock('latex', '\\begin{' .. class .. '}{' .. title .. '}')}
      for _, block in ipairs(el.content) do table.insert(blocks, block) end
      table.insert(blocks, pandoc.RawBlock('latex', '\\end{' .. class .. '}'))
      return blocks
    elseif sizes[class] then
      local blocks = {pandoc.RawBlock('latex', '\\begingroup\\' .. class)}
      for _, block in ipairs(el.content) do table.insert(blocks, block) end
      table.insert(blocks, pandoc.RawBlock('latex', '\\par\\endgroup'))
      return blocks
    end
  end
end


function Table(tbl)
  local simple = pandoc.utils.to_simple_table(tbl)
  local spec = ''
  for _, align in ipairs(simple.aligns) do
    spec = spec .. (align == 'AlignRight' and 'r' or align == 'AlignCenter' and 'c' or 'l')
  end
  local lines = {'\\begin{center}\\begin{tabular}{' .. spec .. '}', '\\toprule'}
  local header = {}
  for _, cell in ipairs(simple.headers) do
    table.insert(header, '\\textcolor{nord8}{\\bfseries ' .. latex(cell) .. '}')
  end
  table.insert(lines, table.concat(header, ' & ') .. ' \\\\')
  table.insert(lines, '\\midrule')
  for _, row in ipairs(simple.rows) do
    local cells = {}
    for _, cell in ipairs(row) do table.insert(cells, latex(cell)) end
    table.insert(lines, table.concat(cells, ' & ') .. ' \\\\')
  end
  table.insert(lines, '\\bottomrule')
  table.insert(lines, '\\end{tabular}\\end{center}')
  return pandoc.RawBlock('latex', table.concat(lines, '\n'))
end

-- Inline code as raw \texttt, so pandoc doesn't mark frames as fragile (fragile frames break with the boxes)
function Code(el)
  return pandoc.RawInline('latex', latex({pandoc.Plain({el})}))
end
