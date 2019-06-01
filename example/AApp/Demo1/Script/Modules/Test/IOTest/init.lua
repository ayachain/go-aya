local exports = {}

exports.WriteTest = function()

    fi = io.open("IOTest.txt", "w")

    if fi == nil then
        print("io.open faild.")
        return false
    end

    fi:write("An exciting, seven-level course that enhances young learners' thinking skills, sharpening their memory while improving their language skills.")
    fi:write("\n")
    fi:write("This exciting seven-level course, from a highly experienced author team, enhances your students' thinking skills.")
    fi:write("\n")
    fi:write(2019)
    fi:write(" ")
    fi:write(6)
    fi:write(" ")
    fi:write(1)
    fi:write("\n")
    fi:write("CreateFileContent Success.\n")
    fi:write("File Content Over")

    fi:close()
    return true

end

exports.ReadTest = function()

    fi = io.open("IOTest.txt", "r")

    if fi == nil then
        print("io.open faild.")
        return false
    end

    line1 = fi:read("l")
    line2 = fi:read("l")
    year = fi:read("n")
    moth = fi:read("n")
    day = fi:read("n")

    fi:read("l")
    line3 = fi:read("l")
    latest = fi:read(17)
    fi:close()

    print("Line1:" ..line1)
    print("Line2:" ..line2)
    print("Year:" ..year)
    print("Moth:" ..moth)
    print("Day:" ..day)
    print("Line3:" ..line3)
    print("Latest:" ..latest)

    return true

end

exports.BaseByIO = function()

    fi = io.open("IOTest_Onput.txt", "w")

    if fi == nil then
        print("io.open faild.")
        return false
    end
    io.output(fi)
    for i = 1,100 do
        io.write("Line" .. i .. "\n")
    end
    io.close(fi)


    fi = io.open("IOTest_Onput.txt", "r")
    io.input(fi)
    for i = 1, 100 do
        print(io.read())
    end
    io.close(fi)

end


exports.TestLine = function()


    local fi = io.open("IOTest_Onput.txt", "r")

    if fi == nil then
        print("io.open faild.")
        return false
    end

    for ct in fi:lines() do
        print(ct)
    end
    fi:close()
    print("fi:lines() running.")


    it = io.lines("IOTest_Onput.txt")

    print("Count 1 " .. it() )
    print("Count 2 " .. it() )

    for ct2 in it do
        print( "It Read :" .. ct2 )
    end

end

exports.Run = function()

    msg = ""

    case1 = exports.WriteTest()
    case2 = exports.ReadTest()

    if case1 and case2 then
        msg = "All Pass"
    else
        if not case1 then msg = msg .. "WriteTest() Failed\n" end
        if not case2 then msg = msg .. "ReadTest() Failed\n" end
    end

    exports.BaseByIO()
    exports.TestLine()

    return msg

end

return exports
