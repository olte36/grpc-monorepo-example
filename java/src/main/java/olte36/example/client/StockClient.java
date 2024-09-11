package olte36.example.client;

import com.google.protobuf.Duration;
import com.google.protobuf.Empty;
import io.grpc.Channel;
import io.grpc.Deadline;
import io.grpc.StatusRuntimeException;
import olte36.example.api.*;

import java.util.Iterator;
import java.util.concurrent.TimeUnit;
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

    public void followStockPrice() {
        FollowRequest request = FollowRequest.newBuilder()
                .setStock(Stock.newBuilder().setTicker("NVDA"))
                .setTrackInterval(Duration.newBuilder().setSeconds(2))
                .build();
        Iterator<FollowResponse> responseIterator = this.blockingStub
                .withDeadline(Deadline.after(10, TimeUnit.SECONDS))
                .follow(request);
        responseIterator.forEachRemaining(s -> {
            System.out.printf("%s -%d\n", s.getStock().getTicker(), s.getPrice());
        });
    }

}
