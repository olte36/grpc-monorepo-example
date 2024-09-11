package olte36.example.server.impl;

import io.grpc.stub.StreamObserver;
import olte36.example.api.TradeRequest;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;

class TradeRequestStreamObserver implements StreamObserver<TradeRequest> {

    private final StockServer server;
    private final Map<String, List<TradeRequest>> orders;
    private final ScheduledFuture<?> executeOrdersTask;

    TradeRequestStreamObserver(StockServer server) {
        this.server = server;
        this.orders = new ConcurrentHashMap<>();
        this.executeOrdersTask = startExecutingOrders();
    }

    @Override
    public void onNext(TradeRequest request) {
        orders.compute(request.getStock().getTicker(), (ticker, ordersForTicker) -> {
            if (ordersForTicker == null) {
                ordersForTicker = new ArrayList<>();
            }
            ordersForTicker.add(request);
            return ordersForTicker;
        });
    }

    @Override
    public void onError(Throwable t) {

    }

    @Override
    public void onCompleted() {
        executeOrdersTask.cancel(false);
    }

    private ScheduledFuture<?> startExecutingOrders() {
        return server.executor.scheduleAtFixedRate(() -> {

        }, 0, 500, TimeUnit.MILLISECONDS);
    }
}
