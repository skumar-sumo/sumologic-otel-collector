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

function process(data)
    if type(data) == 'table' then
      resourceMetrics = data["resourceMetrics"]
      for kResourceMetrics, vResourceMetrics in pairs(resourceMetrics) do
        libraryMetrics = vResourceMetrics["libraryMetrics"]
        for kLibraryMetrics, vLibraryMetrics in pairs(libraryMetrics) do
          metrics = vLibraryMetrics["metrics"]
          for kMetrics, vMetrics in pairs(metrics) do
            dataPoints = vMetrics["sum"]["dataPoints"]
            for kDataPoints, vDataPoints in pairs(dataPoints) do
              dataPoint = vDataPoints
              -- change startTimestamp
              dataPoint["startTimestamp"] = 0
              -- change value
              dataPoint["value"] = 789
              -- add attribute
              dataPoint["attributes"]["lua"] = "true"
            end
          end
        end
      end
    end
    return data
end

