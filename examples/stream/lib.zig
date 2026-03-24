const api = @import("api.zig");
const rt = @import("go2zig_runtime.zig");

pub fn copy_stream(reader: api.GoReader, writer: api.GoWriter) u64 {
    var total: u64 = 0;
    var buf: [4096]u8 = undefined;
    while (true) {
        const n = rt.streamRead(reader, buf[0..]) catch |err| switch (err) {
            error.EndOfStream => break,
            else => @panic("stream read failed"),
        };
        const written = rt.streamWrite(writer, buf[0..n]) catch @panic("stream write failed");
        total += @as(u64, @intCast(written));
    }
    return total;
}
