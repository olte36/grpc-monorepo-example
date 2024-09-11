package olte36.example.server.impl;

import io.grpc.stub.StreamObserver;
import olte36.example.api.OfferRequest;
import olte36.example.api.OfferResponse;
import olte36.example.api.Stock;

class OfferRequestStreamObserver implements StreamObserver<OfferRequest> {

    private final StockServer server;
    private final StreamObserver<OfferResponse> responseObserver;
    private final OfferResponse.Builder responseBuilder;

    OfferRequestStreamObserver(StockServer server, StreamObserver<OfferResponse> responseObserver) {
        this.server = server;
        this.responseObserver = responseObserver;
        this.responseBuilder = OfferResponse.newBuilder();
    }

    @Override
    public void onNext(OfferRequest request) {
        Stock stockToOffer = request.getStock();
        if (server.stocks.containsKey(stockToOffer.getTicker())) {
            Stock rejectedStock = Stock.newBuilder()
                    .setTicker(stockToOffer.getTicker())
                    .setDescription("The stock is already present")
                    .build();
            responseBuilder.addRejectedStocks(rejectedStock);
            return;
        }
        if (request.getInitialPrice() <= 0) {
            Stock rejectedStock = Stock.newBuilder()
                    .setTicker(stockToOffer.getTicker())
                    .setDescription("invalid price")
                    .build();
            responseBuilder.addRejectedStocks(rejectedStock);
            return;
        }
        server.stocks.put(stockToOffer.getTicker(), new InternalStock(
                stockToOffer.getTicker(),
                stockToOffer.getDescription(),
                request.getInitialPrice()
        ));
        responseBuilder.addListedStocks(stockToOffer);
    }

    @Override
    public void onError(Throwable t) {

    }

    @Override
    public void onCompleted() {
        responseObserver.onNext(responseBuilder.build());
        responseObserver.onCompleted();
    }
}
