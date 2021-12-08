function dump(o, indent)
  if type(o) == 'table' then
    local s = ''
    for k,v in pairs(o) do
      if type(k) ~= 'number' then k = '"' .. k .. '"' end
      s = s .. indent .. k .. ': '
      if type(v) == 'table' then
        local st = dump(v, indent .. ' ')
        if string.len(st) > 0 then
          s = s .. '\n' .. st
        end
      else
        s = s .. dump(v, indent) .. '\n'
      end
    end
    return s
  else
    return tostring(o)
  end
end

function process(m)
    print(dump(m, ''))
    return m
end

