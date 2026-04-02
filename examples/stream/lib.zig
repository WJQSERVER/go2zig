const api = @import("api.zig");
const rt = @import("go2zig_runtime.zig");

pub fn copy_stream(reader: api.GoReader, writer: api.GoWriter) api.CopyStreamError!u64 {
    var total: u64 = 0;
    var buf: [65536]u8 = undefined;
    while (true) {
        const n = rt.streamRead(reader, buf[0..]) catch |err| switch (err) {
            error.EndOfStream => break,
            else => return error.StreamReadFailed,
        };
        var pending = buf[0..n];
        while (pending.len > 0) {
            const written = rt.streamWrite(writer, pending) catch return error.StreamWriteFailed;
            if (written == 0) return error.StreamWriteFailed;
            total += @as(u64, @intCast(written));
            pending = pending[written..];
        }
    }
    return total;
}
