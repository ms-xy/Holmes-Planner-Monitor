--[[
Copyright (C) 2016 Holmes Processing.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License version 2 as
published by the Free Software Foundation.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
--]]

-- Chisel description
description = "Prints the CPU workload per second average across all cores"
short_description = "CPU workload average"
category = "CPU Usage"

-- Chisel argument list
args = {}

require "common"
terminal = require "ansiterminal"

islive = false

nproc = 0
fcpu = false
tcpu = 0.0

function printf(s,...)
	if select("#", ...) > 0 then
		return io.write(s:format(...))
	else
		return io.write(s)
	end
end

-- Initialization callback
function on_init()
	nproc = tonumber(io.popen("nproc"):read())
	-- Request the fields we need
	-- thread.cpu = thread.cpu.user + thread.cpu.system
	fcpu = chisel.request_field("thread.cpu")

  -- Filter out all events that aren't procinfo
  chisel.set_filter("evt.type=procinfo")

  return true
end

-- Final chisel initialization
function on_capture_start()
	islive = sysdig.is_live()
	if islive then
		chisel.set_interval_s(1)
	end
	return true
end

-- Event parsing callback
function on_event()
	local cpu = evt.field(fcpu)
	if cpu ~= nil then
		tcpu = tcpu + cpu
	end
	return true
end

-- Periodic timeout callback
function on_interval(ts_s, ts_ns, delta)
	print(tcpu / nproc)
  tcpu = 0.0
  return true
end

-- Called by the engine at the end of the capture (Ctrl-C)
function on_capture_end(ts_s, ts_ns, delta)
	print(tcpu / nproc)
	return true
end
