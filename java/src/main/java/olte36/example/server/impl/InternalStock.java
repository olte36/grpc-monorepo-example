package olte36.example.server.impl;

class InternalStock {
    InternalStock() { }

    InternalStock(String ticker, String description, int price) {
        this.ticker = ticker;
        this.description = description;
        this.price = price;
    }

    String ticker;
    String description;
    int price;
}