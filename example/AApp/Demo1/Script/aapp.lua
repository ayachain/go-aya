print("AAPP:Begin Test ALVM.")

local IOTest = require "Modules.Test.IOTest"
local DBTest = require "Modules.Test.LevelDBTest"

function StartTestIO()
    return IOTest.Run()
end


function StartTestDB()
    return DBTest.Run()
end
