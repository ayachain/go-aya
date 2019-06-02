local exports = {}

exports.WriteTest = function()

    db = adb.open("adb1")

    if db == nil then
        return "adb open failed."
    end

    for i = 1, 1000 do

        k = "key"..tostring(i)
        v = "value"..tostring(i)

        if not db:put( k, v ) then
            return "db:put k:" .. k .. " v:" .. v .. " failed!"
        end

    end

    if not db:close() then
        return "adb close failed"
    end

end

function PrintItInfo( i )
    print( "Key:" .. i:key() .. "\t" .. "Value:" .. i:value() .. "\t" .. "Valid:" .. tostring(i:valid()) .. "\t" .. "Error:" .. tostring(i:error()) )
end

exports.ReadByIterator = function()

    print("ReadByIterator Test Begin.")

    db = adb.open("adb1")
    if db == nil then
        return "adb open failed."
    end

    it = db:newIterator()
    if it == nil then
        db:close()
        return "open iterator failed."
    end

    for i = 1, 10 do
        it:next()
        PrintItInfo(it)
    end

    print("Do prev()\b")
    it:prev()
    PrintItInfo(it)

    print("Do last\b")
    it:last()
    PrintItInfo(it)

    print("Do seek\n")
    it:seek("key999")
    PrintItInfo(it)

    print("Read key200 to key 400\n")
    itlimit = db:newIterator("key200", "key400")
    while itlimit:next() do
        PrintItInfo(itlimit)
    end

    print("Read key990 to latest\n")
    itlimit2 = db:newIterator("key990")
    while itlimit2:next() do
        PrintItInfo(itlimit2)
    end

    print("Read start to key20\n")
    itlimit3 = db:newIterator(nil, "key20")
    while itlimit3:next() do
        PrintItInfo(itlimit3)
    end

    db:close()
    print("AllTest Success.")

end

exports.BatchTest = function()

    print("BatchTest Test Begin.")

    db = adb.open("adb1")
    if db == nil then
        return "adb open failed."
    end

    batch = db:newBatch()
    if batch == nil then
        db:close()
        return "create batch failed."
    end

    batch:put("BatchKey1", "BatchPutValue1")
    batch:put("BatchKey2", "BatchPutValue3")
    batch:put("BatchKey3", "BatchPutValue3")
    batch:delete("BatchKey1")
    if not batch:write() then
        db:close()
        return "batch:write() failed."
    end

    if db:has("BatchKey1") then
        db:close()
        return "batch:delete() failed."
    end

    if not db:has("BatchKey2") then
        db:close()
        return "batch:put() failed."
    end

    print(db:get("BatchKey2"))
    db:close()

    return "Success"

end

exports.Run = function()
    exports.WriteTest()
    exports.ReadByIterator()
    exports.BatchTest()
    return "All Test Case Done."
end

return exports
