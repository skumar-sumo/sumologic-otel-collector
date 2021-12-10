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

function processMetrics(data)
    print('Processing metrics')
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
                        dataPoint["attributes"]["lua"] = {stringVal = "true"}
                    end
                end
            end
        end
    end
    return data
end

function processLogs(data)
    print('Processing logs')
    if type(data) == 'table' then
        resourceLogs = data["resourceLogs"]
        for kResourceLogs, vResourceLogs in pairs(resourceLogs) do
            resource = vResourceLogs["resource"]
            resource["attributes"]["allTheTypes"] = {mapVal = {
                    obviously = {stringVal = "a string value"},
                    coin = {boolVal = false},
                    theAnswer = {intVal = 42},
                    theAnswerEnhanced = {doubleVal = 42.1},
                    everybody = {arrayVal = {{stringVal = "one"}, {stringVal = "two"}, {stringVal = "three"}}},
                    thisIs = {bytesVal = {45, 55, 12}},
            }}
            libraryLogs = vResourceLogs["libraryLogs"]
            for kLibraryLogs, vLibraryLogs in pairs(libraryLogs) do
                logs = vLibraryLogs["logs"]
                for kLogs, vLogs in pairs(logs) do
                    vLogs["attributes"]["lua"] = {stringVal = "true"}
                end
            end
        end
    end
    return data
end

function process(dataType, data)
    print(dump(data, ''))
    if dataType == 'metrics' then
        data = processMetrics(data)
    elseif dataType == 'logs' then 
        data = processLogs(data)
    end
    return data
end

