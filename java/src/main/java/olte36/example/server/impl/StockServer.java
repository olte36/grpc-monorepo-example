package olte36.example.server.impl;

import com.google.protobuf.Empty;
import io.grpc.stub.StreamObserver;
import olte36.example.api.*;

import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicBoolean;

public class StockServer extends StockServiceGrpc.StockServiceImplBase {
    final Map<String, InternalStock> stocks;
    final ScheduledExecutorService executor;

    public StockServer() {
        this.stocks = new ConcurrentHashMap<>();
        this.executor = Executors.newScheduledThreadPool(5);
        initStocks();
    }

    @Override
    public void list(Empty request, StreamObserver<ListResponse> responseObserver) {
        List<Stock> respStocks = stocks.values()
                .stream()
                .map(s -> Stock.newBuilder()
                        .setTicker(s.ticker)
                        .setDescription(s.description)
                        .build()
                ).toList();
        ListResponse resp = ListResponse.newBuilder()
                .addAllStocks(respStocks)
                .build();
        responseObserver.onNext(resp);
        responseObserver.onCompleted();
    }

    @Override
    public StreamObserver<OfferRequest> offer(StreamObserver<OfferResponse> responseObserver) {
        return new OfferRequestStreamObserver(this, responseObserver);
    }

    @Override
    public void follow(FollowRequest request, StreamObserver<FollowResponse> responseObserver) {
        try {
            AtomicBoolean stop = new AtomicBoolean(false);
            Timer timer = new Timer();
            timer.schedule(new TimerTask() {
                @Override
                public void run() {
                    stop.set(true);
                }
            }, 30000);
            while (!stop.get()) {
                InternalStock internalStock = stocks.get(request.getStock().getTicker());
                FollowResponse response;
                synchronized (internalStock) {
                    response = FollowResponse.newBuilder()
                            .setPrice(internalStock.price)
                            .setStock(Stock.newBuilder()
                                    .setTicker(internalStock.ticker)
                            ).build();
                }
                responseObserver.onNext(response);
                //Thread.sleep(request.getTrackInterval().getNanos() / 1000000);
                System.out.println(request.getTrackInterval().getSeconds());
                System.out.println(request.getTrackInterval().getNanos());
                Thread.sleep(2000);
            }
            System.out.println("Stopped");
            responseObserver.onCompleted();
        } catch (InterruptedException e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public StreamObserver<TradeRequest> trade(StreamObserver<TradeResponse> responseObserver) {
        return super.trade(responseObserver);
    }

    private void initStocks() {
        List<InternalStock> stockList = Arrays.asList(
                new InternalStock(
                        "NVDA",
                        "NVIDIA Corporation Common Stock",
                        randPrice()
                ),
                new InternalStock(
                        "PLTR",
                        "Palantir Technologies Inc. Class A Common Stock",
                        randPrice()
                ),
                new InternalStock(
                        "WMT",
                        "Walmart Inc. Common Stock",
                        randPrice()
                )
        );
        stockList.forEach(s -> stocks.put(s.ticker, s));
    }

    private int randPrice() {
        return (new Random().nextInt(40) + 80) * 100;
    }

    private Stock createStock(InternalStock internalStock) {
        return Stock.newBuilder()
                .setTicker(internalStock.ticker)
                .setDescription(internalStock.description)
                .build();
    }
}