const std = @import("std");
const api = @import("api.zig");
const rt = @import("go2zig_runtime.zig");

fn copy_stream_file_handles(reader: api.GoReader, writer: api.GoWriter) api.CopyStreamError!?u64 {
    const src = rt.streamFile(reader) orelse return null;
    const dst = rt.streamFile(writer) orelse return null;

    var threaded: std.Io.Threaded = .init_single_threaded;
    const io = threaded.io();
    var reader_buf: [4096]u8 = undefined;
    var writer_buf: [4096]u8 = undefined;
    var src_reader = src.readerStreaming(io, &reader_buf);
    var dst_writer = dst.writerStreaming(io, &writer_buf);

    const copied = dst_writer.interface.sendFileAll(&src_reader, .unlimited) catch |err| switch (err) {
        error.ReadFailed => return error.StreamReadFailed,
        error.WriteFailed => return error.StreamWriteFailed,
    };
    dst_writer.flush() catch return error.StreamWriteFailed;
    return @intCast(copied);
}

pub fn copy_stream(reader: api.GoReader, writer: api.GoWriter) api.CopyStreamError!u64 {
    if (try copy_stream_file_handles(reader, writer)) |copied| {
        return copied;
    }
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
