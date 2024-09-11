package olte36.example.client;

import com.google.protobuf.Empty;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import olte36.example.api.ListResponse;
import olte36.example.api.StockServiceGrpc;

public final class Main {
    public static void main(String[] args) {
        try {
            ManagedChannel channel = ManagedChannelBuilder.forTarget("localhost:8080")
                    .usePlaintext()
                    .build();
            StockClient client = new StockClient(channel);
            client.printAvailableStocks();
            client.followStockPrice();
        } catch (Exception e) {
            System.out.println(e.getMessage());
            System.exit(1);
        }
    }
}
