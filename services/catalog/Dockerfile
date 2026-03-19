# Stage 1: Build the application
FROM mcr.microsoft.com/dotnet/sdk:10.0-preview AS build

WORKDIR /src

# Copy solution and project files first to leverage Docker cache
COPY Catalog.slnx .
COPY Catalog.Api/Catalog.Api.csproj Catalog.Api/
COPY Catalog.Api.Tests/Catalog.Api.Tests.csproj Catalog.Api.Tests/

RUN dotnet restore

# Copy the rest of the source code
COPY . .

RUN dotnet publish Catalog.Api/Catalog.Api.csproj -c Release -o /app/publish

# Stage 2: Run the application
FROM mcr.microsoft.com/dotnet/aspnet:10.0-preview AS runner

WORKDIR /app

COPY --from=build /app/publish .

EXPOSE 8080

ENV ASPNETCORE_URLS=http://+:8080

ENTRYPOINT ["dotnet", "Catalog.Api.dll"]
