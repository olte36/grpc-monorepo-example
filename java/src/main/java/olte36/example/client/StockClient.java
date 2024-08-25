package olte36.example.client;

import com.google.protobuf.Empty;
import io.grpc.Channel;
import io.grpc.StatusRuntimeException;
import olte36.example.api.ListResponse;
import olte36.example.api.Stock;
import olte36.example.api.StockServiceGrpc;

import java.util.function.Consumer;

public class StockClient {

    private final StockServiceGrpc.StockServiceBlockingStub blockingStub;
    private final StockServiceGrpc.StockServiceStub asyncStub;

    public StockClient(Channel channel) {
        this.blockingStub = StockServiceGrpc.newBlockingStub(channel);
        this.asyncStub = StockServiceGrpc.newStub(channel);
    }

    public void printAvailableStocks() throws Exception {
        try {
            ListResponse resp = blockingStub.list(Empty.newBuilder().build());
            System.out.println("Available stocks:");
            resp.getStocksList().forEach(stock -> {
                try {
                    Thread.sleep(1000);
                    System.out.printf("%s - %s\n", stock.getTicker(), stock.getDescription());
                } catch (InterruptedException e) {
                    throw new RuntimeException(e);
                }
            });
        } catch (StatusRuntimeException e) {
            throw new Exception(String.format("Unable to get the list of stocks, status code: %d", e.getStatus().getCode().value()), e);
        }
    }

}
