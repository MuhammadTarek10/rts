import { ApiProperty } from '@nestjs/swagger';

export enum ResponseStatus {
  SUCCESS = 'success',
  ERROR = 'error',
}

export class ResponseDto<T> {
  @ApiProperty({
    description: 'The data of the response',
    type: Object,
  })
  data: T;

  @ApiProperty({
    description: 'The message of the response',
    example: 'User retrieved successfully',
    type: String,
  })
  message: string;

  @ApiProperty({
    description: 'The error of the response',
    example: 'User not found',
    type: String,
  })
  error?: string;

  @ApiProperty({
    description: 'The status of the response',
    example: ResponseStatus.SUCCESS,
    type: String,
    enum: ResponseStatus,
  })
  status: ResponseStatus;

  constructor(
    data: T,
    message: string,
    status: ResponseStatus,
    error?: string,
  ) {
    this.data = data;
    this.message = message;
    this.status = status;
    if (error) {
      this.error = error;
    }
  }
}
