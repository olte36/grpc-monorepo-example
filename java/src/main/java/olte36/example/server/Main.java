package olte36.example.server;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import olte36.example.server.impl.StockServer;

import java.io.IOException;

public final class Main {
    public static void main(String[] args) {
        try {
            Server server = ServerBuilder.forPort(8080)
                    .addService(new StockServer())
                    .build();
            server.start();
            server.awaitTermination();
        } catch (Exception e) {
            System.out.println(e.getMessage());
            System.exit(1);
        }
    }
}
